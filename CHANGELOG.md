# Changelog

Versioning rule (README): additive = patch, anything a consumer must change for = minor.

## v0.3.0 (unreleased)

- **Prefix-cache block events (feature `kv_block_events`).** `nodeapi.Heartbeat.KVBlocks` carries what changed
  in a worker engine's prefix cache since the last heartbeat — `Stored` / `Removed` block hashes, `Cleared`, a
  `Seq` — and nothing else about a prompt: no token ids, no text. `nodeapi.BlockHash` / `BlockHashes` are the one
  hash both sides call (16 hex of SHA-256 over the parent hash and the block's token ids; its values are pinned by
  a test, because changing it is a protocol change). `nodeapi.PathTokenize` with `TokenizeRequest` /
  `TokenizeResponse` lets a leader, which has no tokenizer, ask a worker's engine for a prompt's token ids.
  `adminapi.PolicyRouting.PrefixBlockWeight` is the knob (0 = off) and `adminapi.Load.PrefixIndex` says whether
  the scorer has anything to score with. All additive; an older reader ignores every one of them.
- **`catalog`: `head_dim` recorded for the two entries whose head dimension is not 128** (`qwen2.5-0.5b-gguf`,
  `llama-3.2-1b`: 64). A control plane sizes a worker's KV cache from `layers × kv_heads × head_dim` and assumes
  128 when the entry says nothing, which doubled the KV term for these two — measured on a card: the 0.5B model's
  footprint is 0.66 GB where the 128 assumption claimed 0.79 GB and 64 gives 0.69 GB.
- **`adminapi.SnapshotKey` loses `RPMLimit`, `TPMLimit`, `AllowedModels` and `QuotaDailyTokens`.** A key is an
  identity — who is calling, until when — and no policy: the leader stopped enforcing per-key rate limits, the
  daily token quota and the model allowlist on 2026-09-28 (per-caller limits are the application layer's, in
  front of the endpoint). A manager that still writes the fields is writing a limit nothing enforces; a leader
  that still reads them ignores them, as JSON does. Minor by the rule above: a consumer that set them must
  stop.

Minor, because consumers are expected to change: the leader and the worker decode the node protocol into
`nodeapi` instead of their own maps and structs, and both sides of the environment contract compare
themselves against `adminapi.Env()`. Nothing that was on the wire before changes — every new type is held
to a golden file that reproduces the body as the binary already wrote it.

- **`adminapi/env.go` — the environment contract as data.** `Env()` (a copy), `Lookup(name)`, the
  `EnvVar{Name, Side, Doc, Since}` row, `SideLeader | SideWorker | SideBoth`, and one `Env*` constant per
  variable so neither side spells a name twice. 25 variables: 5 read by the leader, 12 by the worker, 8
  by both. `Since` is the `Capabilities.Features` key to probe first, or the SDK version that first listed
  a variable that has no feature key; empty = always.
  - The newest rows: `OPOD_OTLP_LOGS_ENDPOINT` (feature `otlp_logs`), and `OPOD_CATALOG_PUBKEY` /
    `OPOD_CATALOG_REQUIRE_SIGNED` (signed catalog directories; no feature key, so `Since` is `v0.3.0`).
- **`adminapi.Retired()`** — `OPOD_UI`, `OPOD_EGRESS`, `OPOD_CALLBACKS`. The dashboard, vendor egress and
  callback sinks they switched off left the binary, so they gate nothing; they are still accepted and
  ignored. A manager that renders them should stop. (`OPOD_PROTOCOLS` is in neither table: it was a word
  in a banner that no code ever parsed.)
- **`nodeapi/` — the node protocol, typed.** Worker → leader: `RegisterRequest` / `RegisterResponse`,
  `Heartbeat` / `HeartbeatResponse`, `Capabilities` + `GPU` (the document inside `hardware_json`, with
  `EncodeHardware` / `DecodeHardware`), `EngineLoad`. Leader → worker: `LoadModelRequest` /
  `LoadModelResponse`, `SleepResponse` (sleep and resume take no body), `Adapter` (an element of
  `OPOD_ADAPTERS`), `LoadAdapterRequest` / `UnloadAdapterRequest` / `AdapterResponse` / `HeldAdapter`,
  `ProcessSpec` / `ProcessInfo` / `StartProcessError` / `StopProcessRequest`, `UploadResponse`. Paths and
  status words are constants.
  - `LoadAdapterRequest.Rank` (`rank`, omitted when zero): the adapter's rank travels on the live add as
    it does in `OPOD_ADAPTERS`, so a worker can refuse an adapter its running engine was not started for.
  - `PathModelUnload` (`/v1/model/unload`, feature `worker_unload`) with `UnloadModelRequest` — the load
    request's source fields without `file` and `pin` — and `UnloadModelResponse`: `unloaded`, `noop`
    (`StatusNoop`; not resident, the call is idempotent) or, with 501, `unsupported` (an engine that cannot
    unload and that the worker did not start). 409 and 502 are plain text, like every other worker error.
  - `Capabilities.Engine` (`Engine` inside `hardware_json`, omitted when empty, appended after the existing
    keys; feature `worker_engine`): the canonical id of the worker's engine driver, so a leader knows what a
    load does there. `Heartbeat.ResidentModels` (`resident_models`, feature of the same name): which of
    `loaded_models` are in memory, from an engine that keeps installed models and loads on request; a
    pointer, because an empty list ("nothing is in memory") is said and only "not stated" is omitted.
    `loaded_models` is documented as what it has always been on such an engine: what the worker answers for.
  - `Heartbeat.LoadedModels`: the comment now says what `null` means to a leader — no report (the engine
    did not answer the worker): liveness advances, nothing recorded about the node's models changes
    (feature `heartbeat_no_report`) — as opposed to `[]`, nothing is loaded. No wire change; goldens untouched.
  - Tests: a golden file per body (`nodeapi/testdata/`), round trip of every golden, unknown keys ignored
    on decode, omitted-versus-zero per field, `hardware_json` byte-exact.

## v0.2.1 (unreleased — committed, not tagged)

Additive.

- `adminapi.PolicyRouting.Revisions` + `RevisionWeight`: weighted routing between plan revisions
  (feature `routing_weights`).
- `adminapi.Load.TTFTP50Ms` / `TTFTP95Ms` and `UsageData.TTFTMS`: time to first token (feature `ttft`).
- `UsageData.CostUSD` is deprecated and always 0; it stays for one additive tag.

## v0.2.0 — 2026-09-14

- `catalog/`: the model catalog (one YAML per model, embedded) and its schema.
- `adminapi.PolicyRouting`: load-aware routing fields (`kvWeight`, `kvSaturationPct`, `prefixAffinity`).

## v0.1.0 — 2026-09-07

- `adminapi/`: the leader's wire types as their own module.
