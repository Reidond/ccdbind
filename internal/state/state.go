package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	Version                int               `json:"version"`
	PinApplied             bool              `json:"pin_applied"`
	OriginalAllowedCPUs    map[string]string `json:"original_allowed_cpus"`
	OSCPUs                 string            `json:"os_cpus"`
	GameCPUs               string            `json:"game_cpus"`
	UpdatedAt              time.Time         `json:"updated_at"`
	LastSuccessfulRestore  time.Time         `json:"last_successful_restore"`
	LastSuccessfulPinApply time.Time         `json:"last_successful_pin_apply"`
}

func DefaultPath() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(base, "ccdbind", "state.json"), nil
}

func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{Version: 1}, nil
		}
		return File{}, err
	}
	var st File
	if err := json.Unmarshal(data, &st); err != nil {
		return File{}, err
	}
	if st.Version == 0 {
		st.Version = 1
	}
	if st.OriginalAllowedCPUs == nil {
		st.OriginalAllowedCPUs = map[string]string{}
	}
	return st, nil
}

func Save(path string, st File) error {
	st.UpdatedAt = time.Now()
	if st.Version == 0 {
		st.Version = 1
	}
	if st.OriginalAllowedCPUs == nil {
		st.OriginalAllowedCPUs = map[string]string{}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
