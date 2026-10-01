# Contributing to opod-sdk

Thanks for your interest. This module is small on purpose: it is the typed contract between an [Opod](https://github.com/opod-io/opod-core) leader and whatever manages it, plus the model catalog. Most contributions are one of three things.

## 1. A catalog entry (the common case)

The catalog is one YAML file per model under [`catalog/`](catalog/). Adding a model is a one-file PR: copy a neighbouring entry, name it `catalog/<id>.yaml`, fill it in against the schema in [`catalog/README.md`](catalog/README.md), and open the PR. No code changes. If you are not sure the model belongs, open a [catalog request](https://github.com/opod-io/opod-sdk/issues/new?template=catalog_request.yml) first.

Check before filing that no entry already exists: `opod model search <name>` on any Opod install, or `ls catalog/`.

## 2. A wire-type fix

The Go types here are **the wire**: the leader marshals them itself, so a JSON tag in this module is a key on the network. If a type does not match what the binary writes, that is a bug here — file it with the JSON the binary produced, or send a PR that changes the type **and** the golden file that froze the old shape.

Rules the review holds to:

- **Additive only within a contract version** (`adminapi.ContractVersion`): a new optional field (`omitempty`), a new row, a new type. Never rename or re-type an existing key. A change to a key already on the wire is a new contract version, and both sides carry the old shape until every deployed binary is past it.
- **Golden files change only when the wire does.** `nodeapi/testdata/*.json` reproduce bodies exactly as the binary wrote them. There is no `-update` flag on purpose: edit the golden by hand, in the same commit as the type, and say in the PR why the wire changed.
- **Standard library only.** Nothing here imports anything outside `std`, and nothing here decides anything. Parsing YAML, HTTP clients and business rules live in the importer.
- A variable the binary stops reading moves from `adminapi.Env()` to `Retired()`; it never disappears.

## 3. Documentation

README, `catalog/README.md`, doc comments and the CHANGELOG. A doc fix needs no issue.

## The loop

```bash
git clone https://github.com/opod-io/opod-sdk
cd opod-sdk
go vet ./... && go test ./...        # this is what CI runs; gofmt -s must be clean too
```

## PR checklist

1. One change per PR.
2. `gofmt -s -l .` prints nothing; `go vet ./...` and `go test ./...` pass.
3. A wire change updates the golden file and a line in [`CHANGELOG.md`](CHANGELOG.md) under the next unreleased tag, classified by the versioning rule in the README (additive = patch, anything a consumer must change for = minor).
4. A catalog entry validates against `catalog/README.md`; the id, licence and sizes are correct.
5. PR title uses the convention of the existing history: `adminapi:`, `nodeapi:`, `catalog:`, `docs:`, `ci:`.
6. Every commit is signed off (`git commit -s`, the [Developer Certificate of Origin](https://developercertificate.org/)). There is no CLA.

## Licence

Apache-2.0 (see [LICENSE](LICENSE)). The sign-off states that you wrote the change or have the right to submit it under that licence.

## Also

- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [SECURITY.md](SECURITY.md) — how to report a vulnerability privately
- [SUPPORT.md](SUPPORT.md) — where to ask
- [`opod-io/opod-core`](https://github.com/opod-io/opod-core) — the runtime that marshals these types; its [CONTRIBUTING](https://github.com/opod-io/opod-core/blob/main/CONTRIBUTING.md) covers engines, routes and the CLI
