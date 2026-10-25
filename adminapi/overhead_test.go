package adminapi

import "testing"

func TestEngineOverheadMiB(t *testing.T) {
	for _, c := range []struct {
		engine, vendor string
		want           int
	}{
		{"vllm", "nvidia", 837},
		{"vllm", "amd", 1103},
		{"vllm", "intel", 1103}, // not measured: the largest
		{"vllm", "", 1103},
		{"sglang", "nvidia", 0}, // no figure: the caller keeps its own
		{"llamacpp", "amd", 0},
	} {
		if got := EngineOverheadMiB(c.engine, c.vendor); got != c.want {
			t.Errorf("EngineOverheadMiB(%q, %q) = %d, want %d", c.engine, c.vendor, got, c.want)
		}
	}
}
