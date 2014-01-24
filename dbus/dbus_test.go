package dbus

import (
	"testing"
)

// TestObjectPath ensures path encoding of the systemd rules works.
func TestObjectPath(t *testing.T) {
	input := "/silly-path/to@a/unit..service"
	output := ObjectPath(input)
	expected := "/silly_2dpath/to_40a/unit_2e_2eservice"

	if string(output) != expected {
		t.Fatalf("Output '%s' did not match expected '%s'", output, expected)
	}
}

// TestNew ensures that New() works without errors.
func TestNew(t *testing.T) {
	_, err := New()

	if err != nil {
		t.Fatal(err)
	}
}
