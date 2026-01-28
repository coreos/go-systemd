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
	"fmt"
	"os/user"
	"regexp"
	"testing"
	"time"

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

			validPath := regexp.MustCompile(`/org/freedesktop/login1/session/[_c][0-9]+`)
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

func TestConn_GetSessionPropertiesContext(t *testing.T) {
	c, err := login1.New()
	if err != nil {
		t.Fatal(err)
	}

	sessions, err := c.ListSessions()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range sessions {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			props, err := c.GetSessionPropertiesContext(ctx, s.Path)
			if err != nil {
				t.Fatal(err)
			}
			if len(props) == 0 {
				t.Fatal("no properties returned")
			}
		}()
	}
}

func TestConn_GetSessionPropertyContext(t *testing.T) {
	c, err := login1.New()
	if err != nil {
		t.Fatal(err)
	}

	sessions, err := c.ListSessions()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range sessions {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			_, err := c.GetSessionPropertyContext(ctx, s.Path, "Remote")
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
}

func TestConn_GetUserPropertiesContext(t *testing.T) {
	c, err := login1.New()
	if err != nil {
		t.Fatal(err)
	}

	users, err := c.ListUsers()
	if err != nil {
		t.Fatal(err)
	}

	for _, u := range users {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			props, err := c.GetUserPropertiesContext(ctx, u.Path)
			if err != nil {
				t.Fatal(err)
			}
			if len(props) == 0 {
				t.Fatal("no properties returned")
			}
		}()
	}
}

func TestConn_GetUserPropertyContext(t *testing.T) {
	c, err := login1.New()
	if err != nil {
		t.Fatal(err)
	}

	users, err := c.ListUsers()
	if err != nil {
		t.Fatal(err)
	}

	for _, u := range users {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			_, err := c.GetUserPropertyContext(ctx, u.Path, "State")
			if err != nil {
				t.Fatal(err)
			}
		}()
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

func Test_Subscribing_to_signals(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("subscribes_to", func(t *testing.T) {
		t.Parallel()

		t.Run("login1_interface", func(t *testing.T) {
			t.Parallel()

			addMatchCalled := false

			connectionWithInterfaceCheck := &mockConnection{
				AddMatchSignalContextF: func(ctx context.Context, options ...dbus.MatchOption) error {
					addMatchCalled = true
					if len(options) < 2 {
						t.Fatalf("Expected at least 2 match options (interface and member)")
					}
					return nil
				},
			}

			conn, err := login1.NewWithConnection(connectionWithInterfaceCheck)
			if err != nil {
				t.Fatalf("Unexpected error creating connection: %v", err)
			}

			if _, err := conn.SubscribeWithContext(ctx, "SessionNew"); err != nil {
				t.Fatalf("Unexpected error subscribing to signals: %v", err)
			}

			if !addMatchCalled {
				t.Fatalf("Expected AddMatchSignalContext to be called")
			}
		})

		t.Run("for_all_given_members", func(t *testing.T) {
			t.Parallel()

			callCount := 0

			connectionWithMemberCheck := &mockConnection{
				AddMatchSignalContextF: func(ctx context.Context, options ...dbus.MatchOption) error {
					callCount++
					return nil
				},
			}

			conn, err := login1.NewWithConnection(connectionWithMemberCheck)
			if err != nil {
				t.Fatalf("Unexpected error creating connection: %v", err)
			}

			expectedMembers := []string{"SessionNew", "SessionRemoved", "UserNew"}
			if _, err := conn.SubscribeWithContext(ctx, expectedMembers...); err != nil {
				t.Fatalf("Unexpected error subscribing to signals: %v", err)
			}

			if callCount != len(expectedMembers) {
				t.Fatalf("Expected AddMatchSignalContext to be called %d times, got %d", len(expectedMembers), callCount)
			}
		})
	})

	t.Run("passes_received_signals_to_channel", func(t *testing.T) {
		t.Parallel()

		signalChannelProvided := false

		connectionWithSignalCheck := &mockConnection{
			SignalF: func(ch chan<- *dbus.Signal) {
				signalChannelProvided = ch != nil
				// Send a test signal to verify the channel works
				go func() {
					ch <- &dbus.Signal{
						Sender: "org.freedesktop.login1",
						Path:   "/org/freedesktop/login1",
						Name:   "org.freedesktop.login1.Manager.SessionNew",
						Body:   []any{"session1", dbus.ObjectPath("/org/freedesktop/login1/session/session1")},
					}
				}()
			},
		}

		conn, err := login1.NewWithConnection(connectionWithSignalCheck)
		if err != nil {
			t.Fatalf("Unexpected error creating connection: %v", err)
		}

		ch, err := conn.SubscribeWithContext(ctx, "SessionNew")
		if err != nil {
			t.Fatalf("Unexpected error subscribing to signals: %v", err)
		}

		if ch == nil {
			t.Fatalf("Expected signal channel to be returned")
		}

		if !signalChannelProvided {
			t.Fatalf("Expected signal channel to be passed to connection")
		}

		// Verify we can receive signals through the channel
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		select {
		case sig := <-ch:
			if sig == nil {
				t.Fatalf("Received nil signal")
			}
			if sig.Name != "org.freedesktop.login1.Manager.SessionNew" {
				t.Fatalf("Expected signal name %q, got %q", "org.freedesktop.login1.Manager.SessionNew", sig.Name)
			}
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for signal")
		}
	})

	t.Run("returns_error_when_adding_match_signal_fails", func(t *testing.T) {
		t.Parallel()

		expectedError := fmt.Errorf("failed to add match")

		connectionWithError := &mockConnection{
			AddMatchSignalContextF: func(ctx context.Context, options ...dbus.MatchOption) error {
				return expectedError
			},
		}

		conn, err := login1.NewWithConnection(connectionWithError)
		if err != nil {
			t.Fatalf("Unexpected error creating connection: %v", err)
		}

		_, err = conn.SubscribeWithContext(ctx, "SessionNew")
		if err == nil {
			t.Fatalf("Expected error when adding match signal fails")
		}
	})
}

// mockConnection is a test helper for mocking dbus.Conn.
type mockConnection struct {
	ObjectF                func(string, dbus.ObjectPath) dbus.BusObject
	AddMatchSignalContextF func(context.Context, ...dbus.MatchOption) error
	SignalF                func(chan<- *dbus.Signal)
}

// AddMatchSignalContext ...
func (m *mockConnection) AddMatchSignalContext(ctx context.Context, options ...dbus.MatchOption) error {
	if m.AddMatchSignalContextF != nil {
		return m.AddMatchSignalContextF(ctx, options...)
	}
	return nil
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
func (m *mockConnection) Signal(ch chan<- *dbus.Signal) {
	if m.SignalF != nil {
		m.SignalF(ch)
	}
}

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

// Connected ...
func (m *mockConnection) Connected() bool {
	return true
}
