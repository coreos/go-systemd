package dbus

import (
	"os"
	"path/filepath"
	"testing"
)

func setupConn(t *testing.T) *Conn {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}

	return conn
}

// Ensure that basic unit starting and stopping works.
func TestStartStopUnit(t *testing.T) {
	target := "/run/systemd/system/start-stop.service"
	conn := setupConn(t)

	// Blindly remove the symlink in case it exists
	err := os.Remove(target)

	// 1. Enable the unit
	abs, err := filepath.Abs("../fixtures/start-stop.service")
	if err != nil {
		t.Fatal(err)
	}

	fixture := []string{abs}

	install, changes, err := conn.EnableUnitFiles(fixture, true, true)

	if install != false {
		t.Fatal("Install was true")
	}

	if len(changes) < 1 {
		t.Fatal("Expected one change, got %v", changes)
	}

	if changes[0].Filename != target {
		t.Fatal("Unexpected target filename")
	}

	// 2. Start the unit
	job, err := conn.StartUnit(filepath.Base(target), "replace")
	if err != nil {
		t.Fatal(err)
	}

	if job != "done" {
		t.Fatal("Job is not done, %v", job)
	}

	units, err := conn.ListUnits()

	var unit *UnitStatus
	for _, u := range units {
		if u.Name == filepath.Base(target) {
			unit = &u
		}
	}

	if unit == nil {
		t.Fatalf("Test unit not found in list")
	}

	if unit.ActiveState != "active" {
		t.Fatalf("Test unit not active")
	}

	// 3. Stop the unit
	job, err = conn.StopUnit(filepath.Base(target), "replace")
	if err != nil {
		t.Fatal(err)
	}

	units, err = conn.ListUnits()

	unit = nil
	for _, u := range units {
		if u.Name == filepath.Base(target) {
			unit = &u
		}
	}

	if unit != nil {
		t.Fatalf("Test unit found in list, should be stopped")
	}
}

// TestGetUnitProperties reads the `-.mount` which should exist on all systemd
// systems and ensures that one of its properties is valid.
func TestGetUnitProperties(t *testing.T) {
	conn := setupConn(t)

	unit := "-.mount"

	info, err := conn.GetUnitProperties(unit)
	if err != nil {
		t.Fatal(err)
	}

	names := info["Wants"].([]string)

	if len(names) < 1 {
		t.Fatal("/ is unwanted")
	}

	if names[0] != "system.slice" {
		t.Fatal("unexpected wants for /")
	}

}
