# opod-sdk

> **The Go contract for [Opod](https://github.com/opod-io/opod-core), the self-hosted, OpenAI-compatible inference runtime: wire types, node protocol, environment contract and model catalog.**

[![Go Reference](https://pkg.go.dev/badge/github.com/opod-io/opod-sdk.svg)](https://pkg.go.dev/github.com/opod-io/opod-sdk)
[![Go](https://img.shields.io/github/go-mod/go-version/opod-io/opod-sdk)](go.mod)
[![Release](https://img.shields.io/github/v/release/opod-io/opod-sdk?sort=semver)](https://github.com/opod-io/opod-sdk/releases/latest)
[![CI](https://github.com/opod-io/opod-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/opod-io/opod-sdk/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/opod-io/opod-sdk?color=blue)](LICENSE)
[![Contributor Covenant](https://img.shields.io/badge/code%20of%20conduct-Contributor%20Covenant%202.1-5e4b8b)](CODE_OF_CONDUCT.md)

[**opod.io**](https://opod.io) · [opod-core](https://github.com/opod-io/opod-core) (the runtime) · [Discussions](https://github.com/opod-io/opod-core/discussions) · [Contributing](CONTRIBUTING.md) · [Support](SUPPORT.md) · Apache-2.0

The typed contract between an **opod** leader (`opod-io/opod-core`, Apache-2.0) and whatever manages it — the opod control plane today, anything else tomorrow. Apache-2.0, stdlib only, no business logic. If you are writing a manager, an autoscaler, a dashboard or a CI check that talks to an Opod leader's `/admin/v1`, or a worker that speaks its node protocol, this module is the one dependency you need; parsing and decisions stay on your side.

```bash
go get github.com/opod-io/opod-sdk@latest
```

| Package | What |
|---|---|
| `adminapi` | The JSON shapes a leader serves on `/admin/v1`, `/loadz` and in the mounted auth / policy snapshot files, plus the frozen route list and feature names. The leader marshals these types itself, so a field here is a field on the wire. Also the **environment contract** (`env.go`): every variable a manager may set on a leader or worker process, as data — `Env()`, `Lookup()`, one `Env*` constant per name, and `Retired()` for variables the binary stopped reading. |
| `nodeapi` | The **node protocol**: the JSON a worker and its leader exchange — register and heartbeat (worker → leader), model load and unload, sleep / resume, LoRA adapters and supervised helper processes (leader → worker) — with the paths as constants. Every type is held to a golden file that reproduces the body as the binary wrote it before the type existed. |
| `catalog` | The model catalog: one YAML per model, embedded (`catalog.FS`, `Names()`, `Read()`) and its schema (`Entry`). The leader and the control plane both read this one set; parsing stays with the importer's YAML library (this module stays stdlib-only). Schema in `catalog/README.md`. |

Planned here next: `connect/` (the client snippet templates `opod connect` and the console share).

```go
import (
	"github.com/opod-io/opod-sdk/adminapi"
	"github.com/opod-io/opod-sdk/nodeapi"
)
```

## The environment contract

`adminapi.Env()` is the list of variables the opod binary declares it reads, each with the side that reads it (`leader`, `worker`, `both`), the leader's own one-line description, and `Since` — the `Capabilities.Features` key to probe before relying on the variable (or, for a variable with no feature key, the SDK version that first listed it; empty = every binary that speaks the contract).

It is held from both sides: the leader asserts that its own declaration equals `Env()`, and a manager asserts that every variable its templates render is in `Env()`. A variable the binary no longer reads moves to `Retired()` rather than disappearing, so a manager that still renders it is told "rendered but retired" instead of "unknown".

The table is what the leader *declares*. The `OPOD_*` spellings of `config.yaml` keys and the launch variables of the container images' entrypoint are not declared as contract data by the leader today, and are not listed.

## The node protocol

`nodeapi` types carry the JSON tags the binary has always written — including the capitalised keys of the two documents that predate tags (`Capabilities` inside `hardware_json`, and `ProcessSpec`). Requests are authenticated by an HMAC over method, path and timestamp; **the body is not signed**, so how a body is serialised (a struct, a map, any key order) changes nothing a receiver verifies. What matters is the key set, omitted-versus-present and value types — which `nodeapi/testdata/*.json` freezes. A golden file changes only when the wire does; there is no `-update` flag on purpose. A receiver ignores keys it does not know, so either side may be the newer one.

## Rules

- **Json tags are the wire.** The leader marshals these types itself; nothing mirrors them.
- **Additive only within a contract version** (`adminapi.ContractVersion`): new optional fields, new rows, new types. Never rename or re-type a key.
- **Standard library only.** Nothing here imports anything else, and nothing here decides anything.

## Versioning

| Change | Tag |
|---|---|
| **Additive** — a new optional field (`omitempty`), a new row in a table, a new type or package, a doc fix. A consumer can take it without touching its code. | patch: `v0.x.Y` |
| **Anything a consumer must change for** — a removed or renamed Go identifier, a field whose type changes, a key that starts or stops being omitted, a variable moved from `Env()` to `Retired()`, a new package a consumer is expected to decode into in place of its own types. | minor: `v0.X.0` |

A change to a key that is already on the wire is neither: it is a new `ContractVersion`, and both sides carry the old shape until every deployed binary is past it.

History: [`CHANGELOG.md`](CHANGELOG.md).

## Contributing, security, licence

- [CONTRIBUTING.md](CONTRIBUTING.md) — a catalog entry is a one-file PR; a wire change comes with its golden file and a CHANGELOG line
- [SECURITY.md](SECURITY.md) — report privately through a GitHub Security Advisory
- [SUPPORT.md](SUPPORT.md) — where each kind of question goes
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) — Contributor Covenant 2.1
- [LICENSE](LICENSE) — Apache-2.0, no CLA, DCO sign-off on every commit
