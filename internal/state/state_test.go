package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPath_UsesXDGStateHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if filepath.Dir(path) != filepath.Join(dir, "ccd-gamed") {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	st := File{Version: 1, PinApplied: true, OriginalAllowedCPUs: map[string]string{"app.slice": ""}}
	if err := Save(path, st); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !loaded.PinApplied {
		t.Fatalf("expected PinApplied true")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected state file to exist: %v", err)
	}
}
