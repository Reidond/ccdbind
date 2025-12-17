package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingFileReturnsDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Interval <= 0 {
		t.Fatalf("expected default interval to be set")
	}
}

func TestLoad_ParsesTOMLAndIgnoreFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	confDir := filepath.Join(dir, "ccd-gamed")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	ignorePath := filepath.Join(confDir, "ignore.txt")
	if err := os.WriteFile(ignorePath, []byte("# comment\nsteam\ncustom-helper\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(ignore): %v", err)
	}

	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`interval = "5s"
env_keys = ["SteamAppId", "STEAM_COMPAT_APP_ID"]
exe_allowlist = ["Foo", "bar"]
pin_session_slice = true
pin_slices = ["app.slice"]
os_cpus = "0-7"
game_cpus = "8-15"
`), 0o644); err != nil {
		t.Fatalf("WriteFile(config): %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got := cfg.Interval.String(); got != "5s" {
		t.Fatalf("interval mismatch: %s", got)
	}
	if !cfg.PinSessionSlice {
		t.Fatalf("expected PinSessionSlice=true")
	}
	if len(cfg.PinSlices) != 1 || cfg.PinSlices[0] != "app.slice" {
		t.Fatalf("unexpected PinSlices: %#v", cfg.PinSlices)
	}
	if cfg.OSCPUsOverride != "0-7" || cfg.GameCPUsOverride != "8-15" {
		t.Fatalf("override mismatch: os=%q game=%q", cfg.OSCPUsOverride, cfg.GameCPUsOverride)
	}
	if !contains(cfg.IgnoreExe, "custom-helper") {
		t.Fatalf("expected ignore list to include ignore.txt entries")
	}
	if !contains(cfg.ExeAllowlist, "foo") {
		t.Fatalf("expected allowlist to be normalized to lower-case")
	}
}

func TestLoad_IgnoreFileWithoutConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	confDir := filepath.Join(dir, "ccd-gamed")
	if err := os.MkdirAll(confDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	ignorePath := filepath.Join(confDir, "ignore.txt")
	if err := os.WriteFile(ignorePath, []byte("custom-helper\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(ignore): %v", err)
	}

	cfg, err := Load(filepath.Join(dir, "missing-config.toml"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !contains(cfg.IgnoreExe, "custom-helper") {
		t.Fatalf("expected ignore list to include ignore.txt entries")
	}
}

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
