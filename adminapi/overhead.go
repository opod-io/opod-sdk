package adminapi

// EngineOverheadMiB is what an engine process holds on its card outside the
// pool its memory fraction bounds (CUDA or HIP context, allocator slack), in
// MiB, per engine and vendor: the one table a manager sizing a budget and a
// worker turning that budget into the engine's fraction both read (PLAN T23.3).
// The figures are measured: vLLM held 837 MiB beyond its pool on an NVIDIA
// card and about 1,103 MiB on an AMD Radeon. A vendor not measured yet takes
// the engine's largest figure, so a budget is never overrun by a guess.
// 0 = no figure for this engine; the caller keeps its own.
func EngineOverheadMiB(engine, vendor string) int {
	byVendor, ok := engineOverheadMiB[engine]
	if !ok {
		return 0
	}
	if mib, ok := byVendor[vendor]; ok {
		return mib
	}
	largest := 0
	for _, mib := range byVendor {
		largest = max(largest, mib)
	}
	return largest
}

var engineOverheadMiB = map[string]map[string]int{
	"vllm": {"nvidia": 837, "amd": 1103},
}
