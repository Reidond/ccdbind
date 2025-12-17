package systemdctl

import "testing"

func TestUnitNameForGameID(t *testing.T) {
	if got := UnitNameForGameID("12345"); got != "game-12345.scope" {
		t.Fatalf("unexpected: %q", got)
	}
	if got := UnitNameForGameID("  "); got != "game-unknown.scope" {
		t.Fatalf("unexpected: %q", got)
	}
	if got := UnitNameForGameID("weird id: (x)"); got != "game-weird_id___x.scope" {
		t.Fatalf("unexpected: %q", got)
	}
}
