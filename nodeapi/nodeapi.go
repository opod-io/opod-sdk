// Package nodeapi is the node protocol: the JSON a worker and its leader
// exchange. It is the second half of the opod wire, beside adminapi (what a
// leader serves to whoever manages it).
//
// Two directions, two listeners:
//
//   - worker → leader, on the leader's listener: the worker registers once
//     (PathRegister), then sends a heartbeat every few seconds
//     (PathHeartbeat) carrying the models its engine holds, whether the engine
//     sleeps, and an engine load sample. Placement is what heartbeats report,
//     never what a request claimed.
//   - leader → worker, on the worker's own listener: make a model resident
//     (PathModelLoad), put the engine to sleep and wake it (PathModelSleep,
//     PathModelResume), attach and detach LoRA adapters (PathAdapters*), and
//     run helper processes — the llama.cpp RPC parts of a sharded model
//     (PathProcess*).
//
// Authentication. Every call carries "X-Opod-Auth: v=1,id=<node>,ts=<unix>,
// sig=<hex>", an HMAC-SHA256 keyed by the node's join token over the string
// "v1\n<METHOD>\n<path without query>\n<ts>". The body is NOT signed — model
// uploads stream gigabytes and are never buffered to be hashed. So the
// signature does not depend on how a body is serialised: key order, and
// whether a struct or a map produced it, change nothing a receiver verifies.
// What a receiver does depend on is the key SET, omitted-versus-present, and
// value types — which is what this package freezes and its golden files hold.
//
// Errors. A worker answers an error with the status code and a plain-text
// body — not JSON — with two exceptions that are typed here: StartProcessError
// (502) and the SleepResponse of an engine with no sleep mode (501). The
// leader answers its own errors in the {"error": {"message", "type"}} shape of
// its /admin/v1 surface.
//
// Rules: json tags are the wire and match, key for key, what the opod binary
// wrote before these types existed; additive only; a receiver ignores keys it
// does not know (held by a test), so either side may be the newer one. This
// package imports nothing but the standard library.
package nodeapi

import (
	"encoding/json"
	"time"
)

// Paths on the leader's listener (worker → leader).
const (
	PathRegister  = "/admin/v1/nodes/register"
	PathHeartbeat = "/admin/v1/nodes/heartbeat"
)

// Paths on the worker's listener (leader → worker).
const (
	PathModelLoad      = "/v1/model/load"
	PathModelSleep     = "/v1/model/sleep"  // POST, no body → SleepResponse
	PathModelResume    = "/v1/model/resume" // POST, no body → SleepResponse
	PathAdapters       = "/v1/adapters"     // GET → []HeldAdapter
	PathAdaptersLoad   = "/v1/adapters/load"
	PathAdaptersUnload = "/v1/adapters/unload"
	PathProcessStart   = "/v1/process/start"  // POST ProcessSpec → 202 ProcessInfo | 502 StartProcessError
	PathProcessStop    = "/v1/process/stop"   // POST StopProcessRequest → 204
	PathProcessList    = "/v1/process/list"   // GET → []ProcessInfo
	PathProcessGet     = "/v1/process/get"    // GET ?id= → ProcessInfo
	PathProcessLogs    = "/v1/process/logs"   // GET ?id=&lines= → text/plain
	PathProcessFile    = "/v1/process/file"   // HEAD ?name=&sha256= → 200 + HeaderFilePath | 404
	PathProcessUpload  = "/v1/process/upload" // POST ?name=&sha256=, raw body → UploadResponse
)

// HeaderFilePath is the response header of PathProcessFile: the absolute
// path of the file on the worker.
const HeaderFilePath = "X-File-Path"

// ---- worker → leader ----

// RegisterRequest is the body of PathRegister. The worker sends every key,
// empty or not.
type RegisterRequest struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	RAMGB    int    `json:"ram_gb"`
	// Address is the host:port of the worker's own listener, as the leader
	// should dial it.
	Address string `json:"address"`
	// HardwareJSON is a Capabilities document encoded as a JSON STRING — JSON
	// inside JSON. The leader stores the string as it arrived and decodes it
	// where it needs a field. EncodeHardware and Hardware do both halves.
	HardwareJSON string `json:"hardware_json"`
	// BootID names the worker PROCESS: minted once per start, sent on register
	// and on every heartbeat, so a leader tells a restarted process from the
	// one it recorded work on (the node id is stable across both). Empty = a
	// worker that predates the field.
	BootID string `json:"boot_id"`
}

// Hardware decodes HardwareJSON. An empty string is the zero Capabilities.
func (r RegisterRequest) Hardware() (Capabilities, error) {
	return DecodeHardware(r.HardwareJSON)
}

// EncodeHardware is the value of RegisterRequest.HardwareJSON for c.
func EncodeHardware(c Capabilities) (string, error) {
	b, err := json.Marshal(c)
	return string(b), err
}

// DecodeHardware reads a hardware_json string (a register body's, or a
// stored node row's). An empty string is the zero Capabilities.
func DecodeHardware(s string) (Capabilities, error) {
	var c Capabilities
	if s == "" {
		return c, nil
	}
	err := json.Unmarshal([]byte(s), &c)
	return c, err
}

// RegisterResponse answers PathRegister.
type RegisterResponse struct {
	Status string `json:"status"` // StatusRegistered
	ID     string `json:"id"`
}

// Capabilities is what a worker knows about its own host and about the
// process it is. The keys are capitalised on the wire: the document predates
// json tags, and node rows that hold it outlive any one binary.
type Capabilities struct {
	Hostname string `json:"Hostname"`
	OS       string `json:"OS"`
	Arch     string `json:"Arch"`
	CPUCores int    `json:"CPUCores"`
	RAMGB    int    `json:"RAMGB"`
	// GPUs is what the worker could detect itself; null when it found none
	// (which includes every card whose tooling the binary cannot probe).
	GPUs []GPU `json:"GPUs"`
	// Role (feature "pd_roles"): "" = a complete server; RolePrefill or
	// RoleDecode = one half of a disaggregated pair.
	Role string `json:"Role,omitempty"`
	// PlanRevision (feature "routing_weights") is the plan revision this
	// worker process was started for; the leader groups workers by it to split
	// traffic between revisions. Zero = not stated.
	PlanRevision int `json:"PlanRevision,omitempty"`
}

// GPU is one detected device.
type GPU struct {
	Name   string `json:"Name"`
	VRAMGB int    `json:"VRAMGB"`
}

// Worker roles (Capabilities.Role).
const (
	RolePrefill = "prefill"
	RoleDecode  = "decode"
)

// Heartbeat is the body of PathHeartbeat.
type Heartbeat struct {
	ID string `json:"id"`
	// LoadedModels is what the engine holds right now, as the ids the worker
	// was asked to load (an adapter as "<base>:<name>"). Always sent; null
	// when the engine did not answer in time.
	LoadedModels []string `json:"loaded_models"`
	BootID       string   `json:"boot_id"`
	// Sleeping is sent only when true (feature "worker_sleep"): the engine
	// dropped its working set, its models are resident but not routable.
	Sleeping bool `json:"sleeping,omitempty"`
	// Load is sent only when the engine reported a sample (feature
	// "load_signals").
	Load *EngineLoad `json:"load,omitempty"`
}

// HeartbeatResponse answers PathHeartbeat. 404 instead means "unknown node —
// register first"; 401/403 mean the token was refused.
type HeartbeatResponse struct {
	Status string `json:"status"` // StatusOK
}

// EngineLoad is one engine's load sample. The leader aggregates the samples
// of a model's live workers into adminapi.Load; time-to-first-token is
// measured by the leader itself and is not part of a sample.
type EngineLoad struct {
	KVUsedPct    float64 `json:"kv_used_pct"`    // KV cache in use, 0–100
	QueueDepth   int64   `json:"queue_depth"`    // requests waiting for a slot
	TokensPerSec float64 `json:"tokens_per_s"`   // generated tokens per second since the previous sample
	PrefixHitPct float64 `json:"prefix_hit_pct"` // prefix-cache hit rate, 0–100 (0 when the engine does not report it)
	SampledAt    int64   `json:"sampled_at"`     // unix seconds
}

// ---- leader → worker: models ----

// LoadModelRequest is the body of PathModelLoad: the catalog SOURCE of a
// model, not a resolved name — the worker resolves the engine-native name
// against its own engine, which may differ from the leader's.
type LoadModelRequest struct {
	ID         string `json:"id"` // required: the catalog id, and the name the model is served under
	OllamaName string `json:"ollama_name"`
	Repo       string `json:"repo"`
	// File is one file inside Repo (a GGUF in a multi-file repository).
	File string `json:"file,omitempty"`
	Path string `json:"path"`
	Pin  bool   `json:"pin"`
}

// LoadModelResponse answers PathModelLoad. "ready" means the load was
// accepted and the engine launched — an engine that loads weights after it
// binds is still warming; residency is what the next heartbeats report.
type LoadModelResponse struct {
	Status string `json:"status"` // StatusReady
	Model  string `json:"model"`  // the engine-native name
}

// SleepResponse answers PathModelSleep and PathModelResume (both take no
// body). An engine without a sleep mode answers 501 with StatusUnsupported
// and a Reason, never a pretended "sleeping"; a leader passes the worker's
// answer through to whoever asked.
type SleepResponse struct {
	Status string `json:"status"` // StatusSleeping | StatusResumed | StatusUnsupported
	Engine string `json:"engine"`
	Reason string `json:"reason,omitempty"`
}

// ---- leader → worker: LoRA adapters (feature "lora") ----

// Adapter is one LoRA adapter of a worker's start-up set: an element of the
// JSON list in the OPOD_ADAPTERS variable (adminapi.EnvAdapters), which a
// manager writes and the worker reads. It is served as "<base>:<name>".
type Adapter struct {
	Name   string `json:"name"`
	Source string `json:"source"` // a Hugging Face repository or a path on the worker
	// Rank is the adapter's own r, when the writer knows it: the engine's LoRA
	// slots are sized at start for the largest rank in the set. Zero = not
	// stated.
	Rank int `json:"rank,omitempty"`
}

// LoadAdapterRequest is the body of PathAdaptersLoad.
type LoadAdapterRequest struct {
	Base   string `json:"base"` // the resident model the adapter attaches to
	Name   string `json:"name"`
	Source string `json:"source"`
}

// UnloadAdapterRequest is the body of PathAdaptersUnload.
type UnloadAdapterRequest struct {
	Base string `json:"base"`
	Name string `json:"name"`
}

// AdapterResponse answers PathAdaptersLoad (StatusReady) and
// PathAdaptersUnload (StatusUnloaded).
type AdapterResponse struct {
	Status string `json:"status"`
	Model  string `json:"model"` // "<base>:<name>"
}

// HeldAdapter is one element of the list GET PathAdapters returns.
type HeldAdapter struct {
	ID     string `json:"id"` // "<base>:<name>"
	Name   string `json:"name"`
	Source string `json:"source"`
}

// ---- leader → worker: helper processes (feature "shards") ----

// ProcessSpec is the body of PathProcessStart: a process the worker
// supervises for the leader. Keys are capitalised on the wire and nothing is
// omitted; the two durations travel as integer nanoseconds.
type ProcessSpec struct {
	ID         string            `json:"ID"`      // caller-assigned, unique on the worker
	Command    string            `json:"Command"` // absolute path or PATH-resolvable binary
	Args       []string          `json:"Args"`
	Env        map[string]string `json:"Env"`        // appended to the worker's own environment
	WorkDir    string            `json:"WorkDir"`    // "" = the worker's
	HealthPort int               `json:"HealthPort"` // > 0: readiness probes this TCP port
	HealthHost string            `json:"HealthHost"` // "" = 127.0.0.1
	LogLines   int               `json:"LogLines"`   // ring-buffer capacity; 0 = the worker's default
	// HealthPath upgrades readiness from "the port is open" to "this path
	// answers 200" — a model server binds before its weights are in.
	HealthPath string `json:"HealthPath"`
	// ReadyTimeout bounds the readiness probe; 0 = the worker's default, which
	// suits a process that listens immediately and not one that loads a model.
	ReadyTimeout time.Duration `json:"ReadyTimeout"`
	// Restart re-launches the process on an abnormal exit, up to MaxRestarts
	// (0 = the worker's default) with RestartBackoff doubling between tries.
	Restart        bool          `json:"Restart"`
	MaxRestarts    int           `json:"MaxRestarts"`
	RestartBackoff time.Duration `json:"RestartBackoff"`
}

// ProcessInfo is the observable state of a supervised process: the 202 of
// PathProcessStart, the answer of PathProcessGet, an element of
// PathProcessList. A start is asynchronous — poll PathProcessGet until Status
// leaves ProcessStarting.
type ProcessInfo struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
	Status    string    `json:"status"` // Process* constants
	ExitErr   string    `json:"exit_err,omitempty"`
	Address   string    `json:"address,omitempty"`
	Restarts  int       `json:"restarts,omitempty"` // automatic restarts so far
}

// StartProcessError is the 502 body of PathProcessStart: the spec was
// refused before anything was spawned (no id, no command, a duplicate id). A
// process that spawns and then fails is a ProcessInfo with ProcessFailed.
type StartProcessError struct {
	Error string       `json:"error"`
	Info  *ProcessInfo `json:"info"` // null when nothing was registered
}

// StopProcessRequest is the body of PathProcessStop.
type StopProcessRequest struct {
	ID string `json:"id"`
}

// UploadResponse answers PathProcessUpload: where the file landed on the
// worker, after its digest matched.
type UploadResponse struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Status words.
const (
	StatusRegistered  = "registered"  // RegisterResponse
	StatusOK          = "ok"          // HeartbeatResponse
	StatusReady       = "ready"       // LoadModelResponse, AdapterResponse (load)
	StatusUnloaded    = "unloaded"    // AdapterResponse (unload)
	StatusSleeping    = "sleeping"    // SleepResponse
	StatusResumed     = "resumed"     // SleepResponse
	StatusUnsupported = "unsupported" // SleepResponse, with 501

	ProcessStarting  = "starting"
	ProcessRunning   = "running"
	ProcessStopped   = "stopped"
	ProcessFailed    = "failed"
	ProcessCrashloop = "crashloop" // Restart was on and MaxRestarts was exceeded
)
