// Package catalog is the model catalog opod ships: one YAML per model (the
// files beside this package, embedded) and the schema they follow. The
// leader (opod-core) and the control plane both read this one set — a model
// added here is known to both, and neither carries a second copy (ADR-040
// gap G12, ROADMAP R9.8). Stdlib only: the files are exposed as an embed.FS
// and parsed by the importer with the YAML library it already has.
package catalog

import (
	"embed"
	"io/fs"
	"sort"
	"strings"
)

//go:embed *.yaml
var files embed.FS

// FS is the embedded catalog: every entry as "<id>.yaml" at the root.
var FS fs.FS = files

// Names lists the embedded files, sorted.
func Names() []string {
	des, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil
	}
	var out []string
	for _, de := range des {
		if !de.IsDir() && strings.HasSuffix(de.Name(), ".yaml") {
			out = append(out, de.Name())
		}
	}
	sort.Strings(out)
	return out
}

// Read returns one embedded file by name.
func Read(name string) ([]byte, error) { return fs.ReadFile(files, name) }

// Entry is the schema of one catalog file. Tags are the wire: yaml for the
// files, json for the APIs that echo an entry.
type Entry struct {
	ID                 string       `yaml:"id"                  json:"id"`
	DisplayName        string       `yaml:"display_name"        json:"display_name"`
	Source             Source       `yaml:"source"              json:"source"`
	SizeBytes          int64        `yaml:"size_bytes"          json:"size_bytes"`
	Quant              string       `yaml:"quant"               json:"quant"`
	ContextWindow      int          `yaml:"context_window"      json:"context_window"`
	Capabilities       []string     `yaml:"capabilities"        json:"capabilities"`
	RecommendedEngines []string     `yaml:"recommended_engines" json:"recommended_engines"`
	Hardware           Hardware     `yaml:"hardware"            json:"hardware"`
	Tags               []string     `yaml:"tags"                json:"tags"`
	Sharding           Sharding     `yaml:"sharding,omitempty"  json:"sharding,omitempty"`
	// Architecture is what a planner validates a split against (tensor
	// parallel divides the heads, pipeline parallel the layers); zero fields
	// = unknown, which warns and never refuses.
	Architecture Architecture `yaml:"architecture,omitempty" json:"architecture,omitempty"`

	License    string `yaml:"license,omitempty"     json:"license,omitempty"`
	LicenseURL string `yaml:"license_url,omitempty" json:"license_url,omitempty"`
	Released   string `yaml:"released,omitempty"    json:"released,omitempty"`

	// Fallback chains the leader walks when this model fails: generic, and
	// per error class.
	Fallback                []string `yaml:"fallback,omitempty"                   json:"fallback,omitempty"`
	FallbackOnContextLength []string `yaml:"fallback_on_context_length,omitempty" json:"fallback_on_context_length,omitempty"`
	FallbackOnContentPolicy []string `yaml:"fallback_on_content_policy,omitempty" json:"fallback_on_content_policy,omitempty"`
}

// Source says where the weights come from.
type Source struct {
	Type       string `yaml:"type"                  json:"type"` // ollama | huggingface | file
	Repo       string `yaml:"repo,omitempty"        json:"repo,omitempty"`
	File       string `yaml:"file,omitempty"        json:"file,omitempty"` // one file within an HF repo (GGUF)
	OllamaName string `yaml:"ollama_name,omitempty" json:"ollama_name,omitempty"`
	Path       string `yaml:"path,omitempty"        json:"path,omitempty"` // a local path (GGUF / safetensors)
}

// Sharding is the llama.cpp-RPC gang default for an entry that needs one.
type Sharding struct {
	Required        bool   `yaml:"required"         json:"required"`
	DefaultShards   int    `yaml:"default_shards"   json:"default_shards"`
	Engine          string `yaml:"engine"           json:"engine"`
	RPCPortBase     int    `yaml:"rpc_port_base"    json:"rpc_port_base"`
	CoordinatorPort int    `yaml:"coordinator_port" json:"coordinator_port"`
}

// Architecture is the model's shape as far as placement needs it.
type Architecture struct {
	Params  int64 `yaml:"params,omitempty"   json:"params,omitempty"`
	Layers  int   `yaml:"layers,omitempty"   json:"layers,omitempty"`
	Heads   int   `yaml:"heads,omitempty"    json:"heads,omitempty"`
	KVHeads int   `yaml:"kv_heads,omitempty" json:"kv_heads,omitempty"`
	Experts int   `yaml:"experts,omitempty"  json:"experts,omitempty"`
}

// Hardware is the entry's stated minimum.
type Hardware struct {
	MinRAMGB  int `yaml:"min_ram_gb"            json:"min_ram_gb"`
	MinVRAMGB int `yaml:"min_vram_gb,omitempty" json:"min_vram_gb,omitempty"`
}
