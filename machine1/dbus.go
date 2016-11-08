/*
Copyright 2015 CoreOS Inc.

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

// Integration with the systemd machined API.  See http://www.freedesktop.org/wiki/Software/systemd/machined/
package machine1

import (
	"fmt"
	"os"
	"strconv"

	"github.com/godbus/dbus"
	"syscall"
)

const (
	dbusInterface = "org.freedesktop.machine1.Manager"
	dbusPath      = "/org/freedesktop/machine1"
	SIGRTMIN      = syscall.Signal(32)
	SIGINT        = syscall.SIGINT
)

// Conn is a connection to systemds dbus endpoint.
type Conn struct {
	conn   *dbus.Conn
	object dbus.BusObject
}

// New() establishes a connection to the system bus and authenticates.
func New() (*Conn, error) {
	c := new(Conn)

	if err := c.initConnection(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Conn) initConnection() error {
	var err error
	c.conn, err = dbus.SystemBusPrivate()
	if err != nil {
		return err
	}

	// Only use EXTERNAL method, and hardcode the uid (not username)
	// to avoid a username lookup (which requires a dynamically linked
	// libc)
	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}

	err = c.conn.Auth(methods)
	if err != nil {
		c.conn.Close()
		return err
	}

	err = c.conn.Hello()
	if err != nil {
		c.conn.Close()
		return err
	}

	c.object = c.conn.Object("org.freedesktop.machine1", dbus.ObjectPath(dbusPath))

	return nil
}

// Close closes the system bus connection
func (c *Conn) Close() error {
	return c.conn.Close()
}

// GetImage may be used to get the image object path for the image with the specified name
func (c *Conn) GetImage(name string) (dbus.ObjectPath, error) {
	var path dbus.ObjectPath
	result := c.object.Call(dbusInterface+".GetImage", 0, name)
	if result.Err != nil {
		return path, result.Err
	}
	pathstr, ok := result.Body[0].(dbus.ObjectPath)
	if ok {
		path = pathstr
	} else {
		return path, fmt.Errorf("Error converting '%v' to dbus.ObjectPath.", result.Body[0])
	}
	return path, nil
}

// GetMachine retrieves the machine object path from dbus
func (c *Conn) GetMachine(name string) (dbus.ObjectPath, error) {
	var path dbus.ObjectPath
	result := c.object.Call(dbusInterface+".GetMachine", 0, name)
	if result.Err != nil {
		return path, result.Err
	}
	pathstr, ok := result.Body[0].(dbus.ObjectPath)
	if ok {
		path = pathstr
	} else {
		return path, fmt.Errorf("Error converting '%v' to dbus.ObjectPath.", result.Body[0])
	}
	return path, nil
}

// GetProperties retrieves the MachineProperties for any given machine
func (c *Conn) GetProperties(name string) (map[string]interface{}, error) {
	var machine map[string]interface{}
	var path dbus.ObjectPath
	var props map[string]dbus.Variant
	result := c.object.Call(dbusInterface+".GetMachine", 0, name)
	if result.Err != nil {
		return machine, result.Err
	}
	pathstr, ok := result.Body[0].(dbus.ObjectPath)
	if ok {
		path = pathstr
	} else {
		return machine, fmt.Errorf("Error converting '%v' to dbus.ObjectPath.", result.Body[0])
	}
	obj := c.conn.Object("org.freedesktop.machine1", path)
	err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "").Store(&props)
	if err != nil {
		return machine, err
	}

	machine = make(map[string]interface{}, len(props))
	for k, v := range props {
		machine[k] = v.Value()
	}
	return machine, nil
}

// KillMachine signals nspawn
func (c *Conn) KillMachine(name string, killsignal syscall.Signal) error {
	return c.object.Call(dbusInterface+".KillMachine", 0, name, "leader", killsignal).Err
}

// PoweroffMachine signals systemd to start the poweroff target
func (c *Conn) PoweroffMachine(name string) error {
	return c.KillMachine(name, SIGRTMIN+4)
}

// RebootMachine signals nspawn to reboot. Works with systemd and sysvinit.
func (c *Conn) RebootMachine(name string) error {
	return c.KillMachine(name, SIGINT)
}

// TerminateMachine terminates the nspawn container
func (c *Conn) TerminateMachine(name string) error {
	return c.object.Call(dbusInterface+".TerminateMachine", 0, name).Err
}

// RegisterMachine registers the container with the systemd-machined
func (c *Conn) RegisterMachine(name string, id []byte, service string, class string, pid int, root_directory string) error {
	return c.object.Call(dbusInterface+".RegisterMachine", 0, name, id, service, class, uint32(pid), root_directory).Err
}
