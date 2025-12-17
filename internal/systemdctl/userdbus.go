package systemdctl

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

type dbusProperty struct {
	Name  string
	Value dbus.Variant
}

type dbusAuxUnit struct {
	Name       string
	Properties []dbusProperty
}

type UserManager struct {
	DryRun bool
	conn   *dbus.Conn
}

func NewUserManager(dryRun bool) (*UserManager, error) {
	if dryRun {
		return &UserManager{DryRun: true}, nil
	}
	conn, err := connectUserBus()
	if err != nil {
		return nil, err
	}
	return &UserManager{conn: conn}, nil
}

func (m *UserManager) Close() error {
	if m.conn != nil {
		m.conn.Close()
	}
	return nil
}

// EnsureTransientScope creates a transient scope (if missing) and attaches PIDs.
// It is safe to call repeatedly.
func (m *UserManager) EnsureTransientScope(ctx context.Context, scopeName string, pids []int, slice string, description string) (created bool, err error) {
	if !strings.HasSuffix(scopeName, ".scope") {
		return false, fmt.Errorf("scope name must end with .scope: %q", scopeName)
	}
	if m.DryRun {
		log.Printf("dry-run: StartTransientUnit(%q) slice=%q pids=%v", scopeName, slice, pids)
		return true, nil
	}
	if m.conn == nil {
		return false, fmt.Errorf("no dbus connection")
	}
	if strings.TrimSpace(slice) == "" {
		slice = "game.slice"
	}

	pidsU32 := make([]uint32, 0, len(pids))
	for _, pid := range pids {
		if pid <= 0 {
			continue
		}
		pidsU32 = append(pidsU32, uint32(pid))
	}

	props := []dbusProperty{
		{Name: "Description", Value: dbus.MakeVariant(description)},
		{Name: "Slice", Value: dbus.MakeVariant(slice)},
		{Name: "PIDs", Value: dbus.MakeVariant(pidsU32)},
	}
	var aux []dbusAuxUnit

	obj := m.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")
	call := obj.CallWithContext(ctx, "org.freedesktop.systemd1.Manager.StartTransientUnit", 0, scopeName, "fail", props, aux)
	if call.Err != nil {
		if isUnitExistsErr(call.Err) {
			return false, nil
		}
		return false, call.Err
	}
	return true, nil
}

// AttachProcessesToUnit attaches the given PIDs to an existing systemd unit.
// The systemd D-Bus signature is: (s unit, s subcgroup, au pids).
func (m *UserManager) AttachProcessesToUnit(ctx context.Context, unit string, subcgroup string, pids []int) error {
	if m.DryRun {
		log.Printf("dry-run: AttachProcessesToUnit(%q, %q) pids=%v", unit, subcgroup, pids)
		return nil
	}
	if m.conn == nil {
		return fmt.Errorf("no dbus connection")
	}
	if len(pids) == 0 {
		return nil
	}
	pidsU32 := make([]uint32, 0, len(pids))
	for _, pid := range pids {
		if pid <= 0 {
			continue
		}
		pidsU32 = append(pidsU32, uint32(pid))
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	obj := m.conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")
	call := obj.CallWithContext(ctx, "org.freedesktop.systemd1.Manager.AttachProcessesToUnit", 0, unit, subcgroup, pidsU32)
	return call.Err
}

func isUnitExistsErr(err error) bool {
	var de dbus.Error
	if errors.As(err, &de) {
		return de.Name == "org.freedesktop.systemd1.UnitExists" || strings.Contains(de.Name, "UnitExists")
	}
	return false
}

func connectUserBus() (*dbus.Conn, error) {
	// First try the standard session bus connection.
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") != "" {
		conn, err := dbus.ConnectSessionBus()
		if err == nil {
			return conn, nil
		}
	}

	uid := os.Getuid()
	addr := ""
	if rt := os.Getenv("XDG_RUNTIME_DIR"); rt != "" {
		addr = "unix:path=" + filepath.Join(rt, "bus")
	} else {
		addr = fmt.Sprintf("unix:path=/run/user/%d/bus", uid)
	}

	conn, err := dbus.Dial(addr)
	if err != nil {
		return nil, err
	}
	if err := conn.Auth(nil); err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.Hello(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
