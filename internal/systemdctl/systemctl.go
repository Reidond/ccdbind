package systemdctl

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type Systemctl struct {
	DryRun bool
}

func (s Systemctl) GetAllowedCPUs(ctx context.Context, unit string) (string, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "--user", "show", "-p", "AllowedCPUs", "--value", unit)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("systemctl show %s: %w (%s)", unit, err, strings.TrimSpace(out.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func (s Systemctl) SetAllowedCPUs(ctx context.Context, unit string, cpus string) error {
	args := []string{"--user", "set-property", "--runtime", unit, fmt.Sprintf("AllowedCPUs=%s", cpus)}
	if s.DryRun {
		log.Printf("dry-run: systemctl %s", strings.Join(args, " "))
		return nil
	}
	cmd := exec.CommandContext(ctx, "systemctl", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl set-property %s: %w (%s)", unit, err, strings.TrimSpace(out.String()))
	}
	return nil
}

func (s Systemctl) StartUnit(ctx context.Context, unit string) error {
	args := []string{"--user", "start", unit}
	if s.DryRun {
		log.Printf("dry-run: systemctl %s", strings.Join(args, " "))
		return nil
	}
	cmd := exec.CommandContext(ctx, "systemctl", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl start %s: %w (%s)", unit, err, strings.TrimSpace(out.String()))
	}
	return nil
}

func DefaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
