package dbus

import (
	"testing"
)

// TestActivation forks out a copy of activation.go example and reads back two
// strings from the pipes that are passed in.
func TestGetUnitProperties(t *testing.T) {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}
	//defer conn.Close()

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
