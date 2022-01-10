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
