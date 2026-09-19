# Changelog

Versioning rule (README): additive = patch, anything a consumer must change for = minor.

## v0.3.0 (unreleased)

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
