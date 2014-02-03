/*
Copyright 2013 CoreOS Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

func setupUnit(target string, conn *Conn, t *testing.T) {
	// Blindly stop the unit in case it is running
	conn.StopUnit(target, "replace")

	// Blindly remove the symlink in case it exists
	targetRun := filepath.Join("/run/systemd/system/", target)
	err := os.Remove(targetRun)

	// 1. Enable the unit
	abs, err := filepath.Abs("../fixtures/" + target)
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

	if changes[0].Filename != targetRun {
		t.Fatal("Unexpected target filename")
	}
}

// Ensure that basic unit starting and stopping works.
func TestStartStopUnit(t *testing.T) {
	target := "start-stop.service"
	conn := setupConn(t)

	setupUnit(target, conn, t)

	// 2. Start the unit
	job, err := conn.StartUnit(target, "replace")
	if err != nil {
		t.Fatal(err)
	}

	if job != "done" {
		t.Fatal("Job is not done, %v", job)
	}

	units, err := conn.ListUnits()

	var unit *UnitStatus
	for _, u := range units {
		if u.Name == target {
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
	job, err = conn.StopUnit(target, "replace")
	if err != nil {
		t.Fatal(err)
	}

	units, err = conn.ListUnits()

	unit = nil
	for _, u := range units {
		if u.Name == target {
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

// TestGetUnitPropertiesRejectsInvalidName attempts to get the properties for a
// unit with an invalid name. This test should be run with --test.timeout set,
// as a fail will manifest as GetUnitProperties hanging indefinitely.
func TestGetUnitPropertiesRejectsInvalidName(t *testing.T) {
	conn := setupConn(t)

	unit := "//invalid#$^/"

	_, err := conn.GetUnitProperties(unit)

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}
