package adminapi

import (
	"encoding/json"
	"strings"
	"testing"
)

// The leader reads the mode under this exact key; a renamed tag would make
// every policy file silently fall back to the default.
func TestPolicyRoutingCacheIsolationWire(t *testing.T) {
	raw, err := json.Marshal(PolicySnapshot{Routing: PolicyRouting{CacheIsolation: CacheIsolationKey}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"cacheIsolation":"key"`) {
		t.Fatalf("wire form %s lacks cacheIsolation", raw)
	}
	var back PolicySnapshot
	if err := json.Unmarshal(raw, &back); err != nil || back.Routing.CacheIsolation != CacheIsolationKey {
		t.Fatalf("round trip: %+v %v", back.Routing, err)
	}
	empty, _ := json.Marshal(PolicyRouting{})
	if strings.Contains(string(empty), "cacheIsolation") {
		t.Fatalf("an unset mode must be omitted, got %s", empty)
	}
}
