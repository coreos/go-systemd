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
	"context"
	"errors"
	"fmt"
	"os/user"
	"regexp"
	"testing"

	"github.com/godbus/dbus/v5"

	"github.com/coreos/go-systemd/v22/login1"
)

// TestNew ensures that New() works without errors.
func TestNew(t *testing.T) {
	_, err := login1.New()
	if err != nil {
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

//nolint:funlen // Many subtests.
func Test_Rebooting_with_context(t *testing.T) {
	t.Parallel()

	t.Run("calls_login1_reboot_method_on_manager_interface", func(t *testing.T) {
		t.Parallel()

		rebootCalled := false

		askForReboot := false

		connectionWithContextCheck := &mockConnection{
			ObjectF: func(string, dbus.ObjectPath) dbus.BusObject {
				return &mockObject{
					CallWithContextF: func(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
						rebootCalled = true

						expectedMethodName := "org.freedesktop.login1.Manager.Reboot"

						if method != expectedMethodName {
							t.Fatalf("Expected method %q being called, got %q", expectedMethodName, method)
						}

						if len(args) != 1 {
							t.Fatalf("Expected one argument to call, got %q", args)
						}

						askedForReboot, ok := args[0].(bool)
						if !ok {
							t.Fatalf("Expected first argument to be of type %T, got %T", askForReboot, args[0])
						}

						if askForReboot != askedForReboot {
							t.Fatalf("Expected argument to be %t, got %t", askForReboot, askedForReboot)
						}

						return &dbus.Call{}
					},
				}
			},
		}

		testConn, err := login1.NewWithConnection(connectionWithContextCheck)
		if err != nil {
			t.Fatalf("Unexpected error creating connection: %v", err)
		}

		if err := testConn.RebootWithContext(context.Background(), askForReboot); err != nil {
			t.Fatalf("Unexpected error rebooting: %v", err)
		}

		if !rebootCalled {
			t.Fatalf("Expected reboot method call on given D-Bus connection")
		}
	})

	t.Run("asks_for_auth_when_requested", func(t *testing.T) {
		t.Parallel()

		rebootCalled := false

		askForReboot := true

		connectionWithContextCheck := &mockConnection{
			ObjectF: func(string, dbus.ObjectPath) dbus.BusObject {
				return &mockObject{
					CallWithContextF: func(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
						rebootCalled = true

						if len(args) != 1 {
							t.Fatalf("Expected one argument to call, got %q", args)
						}

						askedForReboot, ok := args[0].(bool)
						if !ok {
							t.Fatalf("Expected first argument to be of type %T, got %T", askForReboot, args[0])
						}

						if askForReboot != askedForReboot {
							t.Fatalf("Expected argument to be %t, got %t", askForReboot, askedForReboot)
						}

						return &dbus.Call{}
					},
				}
			},
		}

		testConn, err := login1.NewWithConnection(connectionWithContextCheck)
		if err != nil {
			t.Fatalf("Unexpected error creating connection: %v", err)
		}

		if err := testConn.RebootWithContext(context.Background(), askForReboot); err != nil {
			t.Fatalf("Unexpected error rebooting: %v", err)
		}

		if !rebootCalled {
			t.Fatalf("Expected reboot method call on given D-Bus connection")
		}
	})

	t.Run("use_given_context_for_D-Bus_call", func(t *testing.T) {
		t.Parallel()

		testKey := struct{}{}
		expectedValue := "bar"

		ctx := context.WithValue(context.Background(), testKey, expectedValue)

		connectionWithContextCheck := &mockConnection{
			ObjectF: func(string, dbus.ObjectPath) dbus.BusObject {
				return &mockObject{
					CallWithContextF: func(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
						if val := ctx.Value(testKey); val != expectedValue {
							t.Fatalf("Got unexpected context on call")
						}

						return &dbus.Call{}
					},
				}
			},
		}

		testConn, err := login1.NewWithConnection(connectionWithContextCheck)
		if err != nil {
			t.Fatalf("Unexpected error creating connection: %v", err)
		}

		if err := testConn.RebootWithContext(ctx, false); err != nil {
			t.Fatalf("Unexpected error rebooting: %v", err)
		}
	})

	t.Run("returns_error_when_D-Bus_call_fails", func(t *testing.T) {
		t.Parallel()

		expectedError := fmt.Errorf("reboot error")

		connectionWithFailingObjectCall := &mockConnection{
			ObjectF: func(string, dbus.ObjectPath) dbus.BusObject {
				return &mockObject{
					CallWithContextF: func(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
						return &dbus.Call{
							Err: expectedError,
						}
					},
				}
			},
		}

		testConn, err := login1.NewWithConnection(connectionWithFailingObjectCall)
		if err != nil {
			t.Fatalf("Unexpected error creating connection: %v", err)
		}

		if err := testConn.RebootWithContext(context.Background(), false); !errors.Is(err, expectedError) {
			t.Fatalf("Unexpected error rebooting: %v", err)
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

// mockObject is a mock of dbus.BusObject.
type mockObject struct {
	CallWithContextF func(context.Context, string, dbus.Flags, ...interface{}) *dbus.Call
	CallF            func(string, dbus.Flags, ...interface{}) *dbus.Call
}

// mockObject must implement dbus.BusObject to be usable for other packages in tests, though not
// all methods must actually be mockable. See https://github.com/dbus/dbus/issues/252 for details.
var _ dbus.BusObject = &mockObject{}

// CallWithContext ...
//
//nolint:lll // Upstream signature, can't do much with that.
func (m *mockObject) CallWithContext(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	if m.CallWithContextF == nil {
		return &dbus.Call{}
	}

	return m.CallWithContextF(ctx, method, flags, args...)
}

// Call ...
func (m *mockObject) Call(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	if m.CallF == nil {
		return &dbus.Call{}
	}

	return m.CallF(method, flags, args...)
}

// Go ...
func (m *mockObject) Go(method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return &dbus.Call{}
}

// GoWithContext ...
//
//nolint:lll // Upstream signature, can't do much with that.
func (m *mockObject) GoWithContext(ctx context.Context, method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return &dbus.Call{}
}

// AddMatchSignal ...
func (m *mockObject) AddMatchSignal(iface, member string, options ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}

// RemoveMatchSignal ...
func (m *mockObject) RemoveMatchSignal(iface, member string, options ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}

// GetProperty ...
func (m *mockObject) GetProperty(p string) (dbus.Variant, error) { return dbus.Variant{}, nil }

// StoreProperty ...
func (m *mockObject) StoreProperty(p string, value interface{}) error { return nil }

// SetProperty ...
func (m *mockObject) SetProperty(p string, v interface{}) error { return nil }

// Destination ...
func (m *mockObject) Destination() string { return "" }

// Path ...
func (m *mockObject) Path() dbus.ObjectPath { return "" }
