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

// Integration with the systemd D-Bus API.  See http://www.freedesktop.org/wiki/Software/systemd/dbus/
package dbus

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/godbus/dbus"
)

const (
	alpha    = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	num      = `0123456789`
	alphanum = alpha + num
)

// needsEscape checks whether a byte in a potential dbus ObjectPath needs to be escaped
func needsEscape(i int, b byte) bool {
	// Escape everything that is not a-z-A-Z-0-9
	// Also escape 0-9 if it's the first character
	return strings.IndexByte(alphanum, b) == -1 ||
		(i == 0 && strings.IndexByte(num, b) != -1)
}

// PathBusEscape sanitizes a constituent string of a dbus ObjectPath using the
// rules that systemd uses for serializing special characters.
func PathBusEscape(path string) string {
	// Special case the empty string
	if len(path) == 0 {
		return "_"
	}
	n := []byte{}
	for i := 0; i < len(path); i++ {
		c := path[i]
		if needsEscape(i, c) {
			e := fmt.Sprintf("_%x", c)
			n = append(n, []byte(e)...)
		} else {
			n = append(n, c)
		}
	}
	return string(n)
}

// Conn is a connection to systemd's dbus endpoint.
type Conn struct {
	sysconn    *dbus.Conn
	sysobj     *dbus.Object
	jobReqCh   chan *jobRequest
	subscriber struct {
		updateCh chan<- *SubStateUpdate
		errCh    chan<- error
		sync.Mutex
		ignore      map[dbus.ObjectPath]int64
		cleanIgnore int64
	}
}

// New establishes a connection to the system bus and authenticates.
func New() (*Conn, error) {
	return newConnection(dbus.SystemBusPrivate)
}

// NewUserConnection establishes a connection to the session bus and
// authenticates. This can be used to connect to systemd user instances.
func NewUserConnection() (*Conn, error) {
	return newConnection(dbus.SessionBusPrivate)
}

func newConnection(createBus func()(*dbus.Conn, error)) (*Conn, error) {
	c := new(Conn)
	var err error
	c.sysconn, err = createBus()
	if err != nil {
		return nil, err
	}

	// Only use EXTERNAL method, and hardcode the uid (not username)
	// to avoid a username lookup (which requires a dynamically linked
	// libc)
	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}

	err = c.sysconn.Auth(methods)
	if err != nil {
		c.sysconn.Close()
		return nil, err
	}

	err = c.sysconn.Hello()
	if err != nil {
		c.sysconn.Close()
		return nil, err
	}

	c.sysobj = c.sysconn.Object("org.freedesktop.systemd1", dbus.ObjectPath("/org/freedesktop/systemd1"))

	c.jobReqCh = make(chan *jobRequest)
	c.subscriber.ignore = make(map[dbus.ObjectPath]int64)
	c.initDispatch()

	return c, nil
}
