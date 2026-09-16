// Package adminapi is the typed decoupling contract between an opod leader
// and whatever manages it (the control plane today, anything else tomorrow):
// the JSON shapes served on /admin/v1, /loadz and the mounted auth/policy
// snapshot files. The leader marshals these types itself, so a field that
// changes here changes on the wire; the leader and every manager import this
// one package (module github.com/opod-io/opod-sdk), so nothing mirrors it.
//
// Rules: additive only within a contract version; json tags are the wire;
// this package imports nothing but the standard library.
package adminapi

import "time"

// ContractVersion is reported by /admin/v1/version and /admin/v1/capabilities.
const ContractVersion = "v1"

// Route is one method+pattern pair of the frozen /admin/v1 surface
// (chi placeholders exactly as the leader registers them).
type Route struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// Version is GET /admin/v1/version.
type Version struct {
	Version  string `json:"version"`
	Contract string `json:"contract"`
}

// Capabilities is GET /admin/v1/capabilities: the frozen route list plus the
// feature flags a manager may probe instead of sniffing behaviour.
type Capabilities struct {
	Contract string          `json:"contract"`
	Routes   []Route         `json:"routes"`
	Features map[string]bool `json:"features"`
}

// Load is GET /loadz — what the autoscaler reads.
type Load struct {
	PlanRevision    int    `json:"plan_revision"`
	PlanModel       string `json:"plan_model"`
	InFlight        int64  `json:"in_flight"`
	RPM1m           int64  `json:"rpm_1m"`
	Unavailable1m   int64  `json:"unavailable_1m"`
	LastRequestUnix int64  `json:"last_request_unix"`
	TS              int64  `json:"ts"`
	// Engine load signals aggregated over the model's live workers (build
	// item 14; feature "load_signals"): the max KV-cache use, the summed
	// queue, the summed generation rate, the mean prefix-cache hit rate, how
	// many workers are alive and how many of them reported a sample.
	KVUsedPct    float64 `json:"kv_used_pct"`
	QueueDepth   int64   `json:"queue_depth"`
	TokensPerSec float64 `json:"tokens_per_s"`
	PrefixHitPct float64 `json:"prefix_hit_pct"`
	Workers      int     `json:"workers"`
	Reporting    int     `json:"reporting"`
	// Time to the first token a client can use, over the last minute
	// (R15.13; feature "ttft"). A leader that has streamed nothing recently
	// leaves both zero — zero is "not measured", never "instant".
	TTFTP50Ms int64 `json:"ttft_p50_ms,omitempty"`
	TTFTP95Ms int64 `json:"ttft_p95_ms,omitempty"`
}

// UsageData is one usage fact: a request the leader served.
type UsageData struct {
	APIKeyID string `json:"api_key_id"`
	// TTFTMS is the time to the first usable byte of a STREAMED answer
	// (R15.13). 0 on a non-streamed response, where the whole answer arrives
	// at once and latency_ms is the only honest number.
	TTFTMS           int    `json:"ttft_ms,omitempty"`
	UserID           string `json:"user_id"`
	Model            string `json:"model"`
	Protocol         string `json:"protocol"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	LatencyMS        int    `json:"latency_ms"`
	Outcome          string `json:"outcome"`
	// CostUSD is DEPRECATED and always 0: rating and showback are out of this
	// product (ADR-003). It is kept for one additive tag so a client reading it
	// does not break, and will be removed in the next breaking release.
	CostUSD float64 `json:"cost_usd"`
	NodeID  string  `json:"node_id"` // worker that served it ("" = answered locally / never dispatched)
}

// UsageEvent is one row of GET /admin/v1/usage/stream.
type UsageEvent struct {
	ID   int64     `json:"id"`
	Type string    `json:"type"` // "usage"
	TS   time.Time `json:"ts"`
	Data UsageData `json:"data"`
}

// UsageBatch is the page GET /admin/v1/usage/stream returns.
type UsageBatch struct {
	Events []UsageEvent `json:"events"`
	Next   int64        `json:"next"`
	More   bool         `json:"more"`
	Boot   int64        `json:"boot,omitempty"` // leader process start; ids restart with it
}

// LifecycleEvent is one row of GET /admin/v1/events/stream.
type LifecycleEvent struct {
	ID      int64          `json:"id"`
	Type    string         `json:"type"`
	Subject string         `json:"subject"`
	TS      time.Time      `json:"ts"`
	Data    map[string]any `json:"data"`
}

// EventsBatch is the page GET /admin/v1/events/stream returns.
type EventsBatch struct {
	Events []LifecycleEvent `json:"events"`
	Next   int64            `json:"next"`
	More   bool             `json:"more"`
	Boot   int64            `json:"boot,omitempty"`
}

// AuthSnapshot is the auth file a manager mounts (OPOD_AUTH_FILE).
type AuthSnapshot struct {
	Revision    string        `json:"revision"`
	RequireKeys bool          `json:"requireKeys"`
	Keys        []SnapshotKey `json:"keys"`
}

// SnapshotKey is one API key in the auth snapshot (hash, never plaintext).
type SnapshotKey struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Hash          string   `json:"hash"` // sha256 hex of the plaintext key
	Scope         string   `json:"scope,omitempty"`
	RPMLimit      int      `json:"rpmLimit,omitempty"`
	TPMLimit      int      `json:"tpmLimit,omitempty"`
	AllowedModels []string `json:"allowedModels,omitempty"`
	ExpiresAt     string   `json:"expiresAt,omitempty"` // RFC 3339
}

// PolicySnapshot is the policy file a manager mounts (OPOD_POLICY_FILE).
type PolicySnapshot struct {
	Revision   string          `json:"revision"`
	Routing    PolicyRouting   `json:"routing"`
	Logging    PolicyLogging   `json:"logging"`
	Guardrails []GuardrailRule `json:"guardrails"`
}

// PolicyRouting names the fallback target used after the pre-guardrail chain
// when the leader has no capacity (an OpenAI-compatible base URL).
type PolicyRouting struct {
	FallbackURL   string `json:"fallbackUrl,omitempty"`
	FallbackModel string `json:"fallbackModel,omitempty"`
	FallbackKey   string `json:"fallbackKey,omitempty"`
	// Load-aware worker choice (feature "routing_load_aware"): the leader's
	// picker scores a worker as in-flight + queue depth + KVWeight × (KV cache
	// used / 100), from the workers' own heartbeat samples — a full cache
	// counts as KVWeight extra requests; 0 = in-flight only (the old order).
	// A worker at or above KVSaturationPct is placed behind every worker
	// with headroom and receives new requests only when none has any
	// (0 = off). PrefixAffinity pins a prompt prefix to the worker that last
	// served it (its prefix cache), a bounded table with the sticky TTL.
	KVWeight        float64 `json:"kvWeight,omitempty"`
	KVSaturationPct int     `json:"kvSaturationPct,omitempty"`
	PrefixAffinity  bool    `json:"prefixAffinity,omitempty"`
	// Revisions splits traffic between plan revisions (feature
	// "routing_weights", ROADMAP R15.17): each worker registers the plan
	// revision it was started for, and the picker chooses a revision group by
	// weight before it chooses a worker inside it. Empty = no split, which is
	// every endpoint that is not mid-canary.
	//
	// A weight is a share, not a percentage: the leader normalises whatever it
	// is given, so [{1, 9}, {2, 1}] is 10 % on revision 2 and so is
	// [{1, 90}, {2, 10}]. A group with no live worker is skipped and its share
	// goes to the others — a canary whose only pod is restarting must not
	// black-hole its share of the traffic.
	Revisions []RevisionWeight `json:"revisions,omitempty"`
}

// RevisionWeight is one plan revision's share of an endpoint's traffic
// (R15.17). Revision matches what a worker registers as its plan revision.
type RevisionWeight struct {
	Revision int `json:"revision"`
	Weight   int `json:"weight"`
}

// PolicyLogging switches the leader's access log; nil = unchanged.
type PolicyLogging struct {
	AccessLog *bool `json:"accessLog,omitempty"`
}

// GuardrailRule is one webhook guardrail the leader instantiates.
type GuardrailRule struct {
	ID        string            `json:"id"`
	Phase     string            `json:"phase"` // pre | post | logging_only
	URL       string            `json:"url"`
	AuthKey   string            `json:"authKey,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	FailOpen  bool              `json:"failOpen"`
	TimeoutMs int               `json:"timeoutMs,omitempty"`
}
