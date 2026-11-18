package sweep

import "testing"

// isSweepable must match every domain the acceptance tests can create, so a
// crashed run is always reclaimed; it must never match a user's real domain.
func TestIsSweepable(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"tfacc-test-123.example.com", true},
		{"tf-acc-test.example.com", true},
		{"forwardemailprovider.com", true}, // acceptance-test fixture domain
		{"akbnet.org", false},              // real user domain
		{"example.com", false},
		{"tfacc", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSweepable(tt.name); got != tt.want {
				t.Errorf("isSweepable(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
