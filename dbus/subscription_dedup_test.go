package dbus

import "testing"

func TestNoteSubStateLockedDedup(t *testing.T) {
	c := &Conn{}
	c.subStateSubscriber.last = make(map[string]string)

	if !c.noteSubStateLocked("hello.service", "running") {
		t.Fatal("first emission for a unit must be allowed")
	}
	if c.noteSubStateLocked("hello.service", "running") {
		t.Fatal("identical consecutive SubState must be suppressed")
	}
	if !c.noteSubStateLocked("hello.service", "dead") {
		t.Fatal("SubState change must be emitted")
	}
	if c.noteSubStateLocked("hello.service", "dead") {
		t.Fatal("repeated dead must be suppressed")
	}
	if !c.noteSubStateLocked("other.service", "dead") {
		t.Fatal("different unit must not share suppression")
	}
}
