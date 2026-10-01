# Getting help with opod-sdk

This module is the typed contract between an [Opod](https://github.com/opod-io/opod-core) leader and whatever manages it: wire types, the node protocol, the environment contract and the model catalog. Questions about *running* Opod belong in the core repo; questions about *these types* belong here.

| You want to… | Go to |
|---|---|
| Ask how to use a type, a route or the catalog from your own manager | [Discussions in opod-core](https://github.com/opod-io/opod-core/discussions) — one community for both repos |
| Report a type that does not match what the leader writes | [Bug report](https://github.com/opod-io/opod-sdk/issues/new?template=bug_report.yml) — paste the JSON the binary produced |
| Add or fix a model in the catalog | [Catalog request](https://github.com/opod-io/opod-sdk/issues/new?template=catalog_request.yml), or a PR adding `catalog/<id>.yaml` (schema in [`catalog/README.md`](catalog/README.md)) |
| Report a security problem | **Privately**, per [SECURITY.md](SECURITY.md) |
| Ask about the runtime itself (install, engines, routing) | [opod-core](https://github.com/opod-io/opod-core) |

## Read first

- [README](README.md) — the packages, the environment contract, the node protocol, the versioning rule
- [CHANGELOG](CHANGELOG.md) — every tag and what a consumer must change for
- [Go package docs](https://pkg.go.dev/github.com/opod-io/opod-sdk)
