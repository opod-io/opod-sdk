package adminapi

// The process-environment contract: every variable a manager of an opod
// process (any manager — a control plane, a systemd unit, a script) may set
// to steer a leader or a worker, as data.
//
// The leader declares this list itself (its config package parses exactly
// these names, once, and nothing else in its library code reads the
// environment); this file is the same list for the other side of the wall, so
// a manager that writes a pod template and the binary that reads it spell
// each name once, here. Two drift tests hold it: the leader asserts its own
// declaration equals Env(), and a manager asserts that every variable it
// renders is in Env() — or in Retired(), which it should then stop rendering.
//
// What is NOT in the table, on purpose:
//
//   - Configuration overrides. The opod binary also accepts an OPOD_*
//     spelling of most config.yaml keys (listen address, data directory, TLS
//     pair, join token …), and the container images' entrypoint reads a few
//     launch variables of its own (role, leader URL, the model to load).
//     The leader does not declare those as contract data today, so they are
//     not listed here: this table is what the leader declares, not what
//     someone found by reading its source.
//   - OPOD_PROTOCOLS. It was named in a start-up banner and parsed by no
//     code, ever. It was never a variable, so it is neither in Env() nor in
//     Retired().
//
// Rules: additive within a contract version. A variable the binary stops
// reading moves to Retired() — it is never silently deleted, because a
// manager still rendering it needs a better answer than "unknown".

// The side of the leader↔worker pair that reads a variable.
const (
	SideLeader = "leader"
	SideWorker = "worker"
	SideBoth   = "both"
)

// Names of the environment contract. Every constant is a row of Env() and
// every row has a constant (held by a test).
const (
	// leader
	EnvPlanFile         = "OPOD_PLAN_FILE"
	EnvAuthFile         = "OPOD_AUTH_FILE"
	EnvPolicyFile       = "OPOD_POLICY_FILE"
	EnvCoordinatorNode  = "OPOD_COORDINATOR_NODE"
	EnvOTLPLogsEndpoint = "OPOD_OTLP_LOGS_ENDPOINT"
	// worker
	EnvAccelerator   = "OPOD_ACCELERATOR"
	EnvEngineFlags   = "OPOD_ENGINE_FLAGS"
	EnvAdapters      = "OPOD_ADAPTERS"
	EnvRejectBearer  = "OPOD_REJECT_BEARER"
	EnvSleepMode     = "OPOD_SLEEP_MODE"
	EnvWorkerRole    = "OPOD_WORKER_ROLE"
	EnvPlanRevision  = "OPOD_PLAN_REVISION"
	EnvAdvertiseAddr = "OPOD_ADVERTISE_ADDR"
	EnvNodeID        = "OPOD_NODE_ID"
	EnvLeaderCA      = "OPOD_LEADER_CA"
	EnvVRAMBudgetGB  = "OPOD_VRAM_BUDGET_GB"
	EnvGPUIndex      = "OPOD_GPU_INDEX"
	// both
	EnvCatalogDir           = "OPOD_CATALOG_DIR"
	EnvCatalogPubKey        = "OPOD_CATALOG_PUBKEY"
	EnvCatalogRequireSigned = "OPOD_CATALOG_REQUIRE_SIGNED"
	EnvSkipSourceCheck      = "OPOD_SKIP_SOURCE_CHECK"
	EnvHFToken              = "HF_TOKEN"
	EnvHFEndpoint           = "HF_ENDPOINT"
	EnvModelRevision        = "OPOD_MODEL_REVISION"
	EnvModelSHA256          = "OPOD_MODEL_SHA256"
)

// Names that were part of the contract and are not any more (Retired()).
const (
	RetiredEnvUI        = "OPOD_UI"
	RetiredEnvEgress    = "OPOD_EGRESS"
	RetiredEnvCallbacks = "OPOD_CALLBACKS"
)

// EnvVar is one row of the environment contract.
type EnvVar struct {
	Name string `json:"name"` // the variable
	Side string `json:"side"` // SideLeader | SideWorker | SideBoth: who reads it
	Doc  string `json:"doc"`  // one line, the leader's own wording
	// Since says which binaries read the variable: a feature key of
	// Capabilities.Features (probe it before relying on the variable), or —
	// for a variable with no feature key — the SDK version that first listed
	// it ("v0.3.0"). Empty = every binary that speaks this contract version.
	Since string `json:"since,omitempty"`
}

// RetiredEnvVar is a variable the binary no longer reads. Setting one is
// harmless — it is accepted and ignored — and pointless.
type RetiredEnvVar struct {
	Name      string `json:"name"`
	RetiredIn string `json:"retired_in"` // the SDK version that records the retirement
	Note      string `json:"note"`       // what it used to do and why nothing replaces it
}

// envTable is the contract, in the leader's own order (leader, worker, both).
// Doc strings are the leader's, verbatim, so the leader's drift test can
// compare whole rows.
var envTable = []EnvVar{
	// leader
	{Name: EnvPlanFile, Side: SideLeader, Since: "plan_file", Doc: `the mounted plan the leader serves within (default /etc/opod/plan.json; absent = standalone)`},
	{Name: EnvAuthFile, Side: SideLeader, Since: "auth_file", Doc: `the mounted auth snapshot: keys + requireKeys (default /etc/opod-auth/auth.json; "off" = no watcher)`},
	{Name: EnvPolicyFile, Side: SideLeader, Since: "policy_file", Doc: `the mounted policy snapshot: routing, logging, guardrails (default /etc/opod-auth/policy.json; "off" = no watcher)`},
	{Name: EnvCoordinatorNode, Side: SideLeader, Since: "shards", Doc: `pin the llama.cpp RPC coordinator to a node id ("local" = the leader itself)`},
	{Name: EnvOTLPLogsEndpoint, Side: SideLeader, Since: "otlp_logs", Doc: `OTLP/HTTP collector for the leader's own log records (URL or host:port); stderr keeps working, the queue is bounded and never blocks. Empty = off`},
	// worker
	{Name: EnvAccelerator, Side: SideWorker, Doc: `the vendor the manager placed the worker on (nvidia | amd | intel | tt | none); unset = what the worker detects`},
	{Name: EnvEngineFlags, Side: SideWorker, Doc: `JSON map of engine flags from the plan (tp, max_model_len, ctx, ngl, …)`},
	{Name: EnvAdapters, Side: SideWorker, Since: "lora", Doc: `JSON list of LoRA adapters [{name, source, rank}] served as <model>:<name>`},
	{Name: EnvRejectBearer, Side: SideWorker, Doc: `1 = the worker's API accepts HMAC only, never a bearer token`},
	{Name: EnvSleepMode, Side: SideWorker, Since: "worker_sleep", Doc: `1 = vLLM starts with sleep mode on (the sleep autoscale tier)`},
	{Name: EnvWorkerRole, Side: SideWorker, Since: "pd_roles", Doc: `prefill | decode for disaggregated serving; unset = a whole worker`},
	{Name: EnvPlanRevision, Side: SideWorker, Since: "routing_weights", Doc: `the plan revision this worker process was started for; the leader routes a share of traffic per revision (R15.17)`},
	{Name: EnvAdvertiseAddr, Side: SideWorker, Doc: `the host:port the leader should dial (overlay / multi-NIC hosts)`},
	{Name: EnvNodeID, Side: SideWorker, Doc: `a stable node id across restarts (else node.yaml, else random)`},
	{Name: EnvLeaderCA, Side: SideWorker, Since: "tls_listener", Doc: `PEM certificate the worker trusts for a TLS leader (exactly that one)`},
	{Name: EnvVRAMBudgetGB, Side: SideWorker, Since: "vram_budget", Doc: `the slice of the card this worker may use (shared placement); read by the engine launch`},
	{Name: EnvGPUIndex, Side: SideWorker, Since: "vram_budget", Doc: `the device index the worker was pinned to (informational; the engine sees CUDA_VISIBLE_DEVICES & co)`},
	// both
	{Name: EnvCatalogDir, Side: SideBoth, Doc: `a catalog directory that overrides the bundled entries (beaten only by ~/.opod/catalog)`},
	{Name: EnvCatalogPubKey, Side: SideBoth, Since: "v0.3.0", Doc: `a minisign public key (the base64 line, or a file holding one): a catalog file in a directory that has a <file>.minisig beside it must verify against it, and a signature that does not is always a refusal. The embedded catalog is never checked. Empty = signatures are ignored`},
	{Name: EnvCatalogRequireSigned, Side: SideBoth, Since: "v0.3.0", Doc: `1 = a catalog file in a directory with NO signature is refused too (needs OPOD_CATALOG_PUBKEY); default: unsigned files load`},
	{Name: EnvSkipSourceCheck, Side: SideBoth, Doc: `1 = never HEAD-check a model's upstream (air-gapped mirrors)`},
	{Name: EnvHFToken, Side: SideBoth, Doc: `Hugging Face token for gated repositories`},
	{Name: EnvHFEndpoint, Side: SideBoth, Doc: `Hugging Face Hub base URL (a mirror); default https://huggingface.co`},
	{Name: EnvModelRevision, Side: SideBoth, Since: "model_revision", Doc: `pin the model's Hub revision (commit sha, tag or branch); cached under <repo>@<rev> so two revisions coexist. Empty = main, which moves`},
	{Name: EnvModelSHA256, Side: SideBoth, Since: "model_revision", Doc: `the expected sha256 of the model file; a mismatch removes the file and fails the load`},
}

// retiredTable: the three surface switches. The dashboard, vendor egress and
// callback sinks they turned off left the binary, so nothing was left to
// gate; the binary neither reads nor warns about them.
var retiredTable = []RetiredEnvVar{
	{Name: RetiredEnvUI, RetiredIn: "v0.3.0", Note: `switched the built-in web dashboard off; the binary has no dashboard, so there is nothing to switch. Accepted and ignored`},
	{Name: RetiredEnvEgress, RetiredIn: "v0.3.0", Note: `switched outbound calls to hosted model vendors off; the binary makes none. Accepted and ignored`},
	{Name: RetiredEnvCallbacks, RetiredIn: "v0.3.0", Note: `switched the callback sinks off; the binary has none. Accepted and ignored`},
}

// Env returns the environment contract, in the leader's order. The slice is a
// copy: changing it does not change the package's table.
func Env() []EnvVar {
	return append([]EnvVar(nil), envTable...)
}

// Lookup returns the contract row for a variable name.
func Lookup(name string) (EnvVar, bool) {
	for _, v := range envTable {
		if v.Name == name {
			return v, true
		}
	}
	return EnvVar{}, false
}

// Retired returns the variables that left the contract (a copy). A manager
// that still renders one of these can say "rendered but retired" instead of
// "unknown", and drop it at its next template change.
func Retired() []RetiredEnvVar {
	return append([]RetiredEnvVar(nil), retiredTable...)
}

// LookupRetired returns the retirement row for a variable name.
func LookupRetired(name string) (RetiredEnvVar, bool) {
	for _, v := range retiredTable {
		if v.Name == name {
			return v, true
		}
	}
	return RetiredEnvVar{}, false
}
