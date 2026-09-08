# opod-sdk

The typed contract between an **opod** leader (`opod-io/opod-core`, Apache-2.0) and whatever manages it — the opod control plane today, anything else tomorrow. Apache-2.0, stdlib only, no business logic.

| Package | What |
|---|---|
| `adminapi` | The JSON shapes a leader serves on `/admin/v1`, `/loadz` and in the mounted auth / policy snapshot files, plus the frozen route list and feature names. The leader marshals these types itself, so a field here is a field on the wire. |

Planned here next: `catalog/` (the model catalog schema and YAMLs) and `connect/` (the client snippet templates `opod connect` and the console share).

Rules: additive only within a contract version (`adminapi.ContractVersion`); json tags are the wire; nothing here imports anything but the standard library.

```go
import "github.com/opod-io/opod-sdk/adminapi"
```
