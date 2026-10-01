# Security Policy

## What this module is

`opod-sdk` is types and data: Go structs with JSON tags, route constants, an environment-variable table and YAML catalog entries. It runs no server and makes no network calls. A security defect here is therefore almost always one of:

- a type whose shape lets a manager or a leader misread a security-relevant field (a key's expiry, a scope, a revocation);
- a catalog entry that points at the wrong artifact (a wrong repository, a wrong checksum field, a licence misstated);
- a documented route or variable that leaks something it should not.

The code that *acts* on these types lives in [`opod-io/opod-core`](https://github.com/opod-io/opod-core). If you are unsure which repository a finding belongs to, report it here and we will route it. Do not file it twice.

## Supported versions

Only the latest tagged release is supported; the module is pre-1.0 and fixes land on `main` and in the next tag. A consumer pins by tag (`go get github.com/opod-io/opod-sdk@latest`).

## Reporting a vulnerability

**Do not file a public issue for security bugs.**

Preferred: open a private GitHub Security Advisory at https://github.com/opod-io/opod-sdk/security/advisories/new. Alternatively email `hadi.work.ca@gmail.com`.

Include a description and impact, the type or catalog file concerned, and the JSON or YAML that demonstrates it.

## Timeline

- Acknowledgement within 48 hours
- Triage and severity within 7 days
- Fix on `main` and a tag within 30 days
- Public advisory, coordinated with the reporter, within 90 days

## Scope

In scope: everything in this repository and the module published at `github.com/opod-io/opod-sdk`.

Out of scope: the Opod binary and its images (report to [opod-core](https://github.com/opod-io/opod-core/security/policy)); the upstream model weights the catalog points at (report to their publishers).
