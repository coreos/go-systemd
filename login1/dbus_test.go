// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package login1_test

import (
	"fmt"
	"os/user"
	"regexp"
	"testing"

	"github.com/godbus/dbus/v5"

	"github.com/coreos/go-systemd/v22/login1"
)

// TestNew ensures that New() works without errors.
func TestNew(t *testing.T) {
	if _, err := login1.New(); err != nil {
		t.Fatal(err)
	}
}

func TestListSessions(t *testing.T) {
	c, err := login1.New()
	if err != nil {
		t.Fatal(err)
	}

	sessions, err := c.ListSessions()
	if err != nil {
		t.Fatal(err)
	}

	if len(sessions) > 0 {
		for _, s := range sessions {
			lookup, err := user.Lookup(s.User)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(s.UID) != lookup.Uid {
				t.Fatalf("expected uid '%d' but got '%s'", s.UID, lookup.Uid)
			}

			validPath := regexp.MustCompile(`/org/freedesktop/login1/session/_[0-9]+`)
			if !validPath.MatchString(fmt.Sprint(s.Path)) {
				t.Fatalf("invalid session path: %s", s.Path)
			}
		}
	}
}

func TestListUsers(t *testing.T) {
	c, err := login1.New()
	if err != nil {
		t.Fatal(err)
	}

	users, err := c.ListUsers()
	if err != nil {
		t.Fatal(err)
	}

	if len(users) > 0 {
		for _, u := range users {
			lookup, err := user.Lookup(u.Name)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(u.UID) != lookup.Uid {
				t.Fatalf("expected uid '%d' but got '%s'", u.UID, lookup.Uid)
			}

			validPath := regexp.MustCompile(`/org/freedesktop/login1/user/_[0-9]+`)
			if !validPath.MatchString(fmt.Sprint(u.Path)) {
				t.Fatalf("invalid user path: %s", u.Path)
			}
		}
	}
}

func Test_Creating_new_connection_with_custom_connection(t *testing.T) {
	t.Parallel()

	t.Run("connects_to_global_login1_path_and_interface", func(t *testing.T) {
		t.Parallel()

		objectConstructorCalled := false

		connectionWithContextCheck := &mockConnection{
			ObjectF: func(dest string, path dbus.ObjectPath) dbus.BusObject {
				objectConstructorCalled = true

				expectedDest := "org.freedesktop.login1"

				if dest != expectedDest {
					t.Fatalf("Expected D-Bus destination %q, got %q", expectedDest, dest)
				}

				expectedPath := dbus.ObjectPath("/org/freedesktop/login1")

				if path != expectedPath {
					t.Fatalf("Expected D-Bus path %q, got %q", expectedPath, path)
				}

				return nil
			},
		}

		if _, err := login1.NewWithConnection(connectionWithContextCheck); err != nil {
			t.Fatalf("Unexpected error creating connection: %v", err)
		}

		if !objectConstructorCalled {
			t.Fatalf("Expected object constructor to be called")
		}
	})

	t.Run("returns_error_when_no_custom_connection_is_given", func(t *testing.T) {
		t.Parallel()

		testConn, err := login1.NewWithConnection(nil)
		if err == nil {
			t.Fatalf("Expected error creating connection with no connector")
		}

		if testConn != nil {
			t.Fatalf("Expected connection to be nil when New returns error")
		}
	})
}

// mockConnection is a test helper for mocking dbus.Conn.
type mockConnection struct {
	ObjectF func(string, dbus.ObjectPath) dbus.BusObject
}

// Auth ...
func (m *mockConnection) Auth(authMethods []dbus.Auth) error {
	return nil
}

// Hello ...
func (m *mockConnection) Hello() error {
	return nil
}

// Signal ...
func (m *mockConnection) Signal(ch chan<- *dbus.Signal) {}

// Object ...
func (m *mockConnection) Object(dest string, path dbus.ObjectPath) dbus.BusObject {
	if m.ObjectF == nil {
		return nil
	}

	return m.ObjectF(dest, path)
}

// Close ...
func (m *mockConnection) Close() error {
	return nil
}

// BusObject ...
func (m *mockConnection) BusObject() dbus.BusObject {
	return nil
}

// Connected ...
func (m *mockConnection) Connected() bool {
	return true
}
