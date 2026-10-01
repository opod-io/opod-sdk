<!-- Thanks for contributing. One change per PR; see CONTRIBUTING.md. -->

## Summary

<!-- What changes and why. Link the issue if there is one. -->

Closes #

## Kind of change

- [ ] Catalog entry (`catalog/<id>.yaml` — no code)
- [ ] Wire type — **additive** (new `omitempty` field, new row, new type) → patch tag
- [ ] Wire type — **a consumer must change for it** (removed/renamed identifier, re-typed field, a key that starts or stops being omitted, `Env()` → `Retired()`) → minor tag
- [ ] Docs / CI

## For a wire change

- [ ] The golden file under `nodeapi/testdata/` (or the adminapi fixture) changed in this commit, and the PR says why the wire changed
- [ ] A line in `CHANGELOG.md` under the next unreleased tag, classified by the rule above
- [ ] Nothing new is imported outside the standard library

## For a catalog entry

- [ ] Validates against `catalog/README.md`; id, publisher, licence, sizes and context window are correct
- [ ] No duplicate of an existing entry (`ls catalog/`)

## Checks

- [ ] `gofmt -s -l .` prints nothing · `go vet ./...` · `go test ./...`
- [ ] Commits are signed off (`git commit -s`)
