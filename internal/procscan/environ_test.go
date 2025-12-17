package procscan

import "testing"

func TestToSetLower(t *testing.T) {
	set := toSetLower([]string{" a ", "", "A"})
	if len(set) != 1 {
		t.Fatalf("expected 1, got %d", len(set))
	}
}
