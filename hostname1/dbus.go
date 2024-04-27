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

// Package hostname1 provides integration with the systemd hostnamed API.
// See https://www.freedesktop.org/software/systemd/man/latest/org.freedesktop.hostname1.html
package hostname1

import (
	"os"
	"strconv"

	"github.com/godbus/dbus/v5"
)

const (
	dbusDest = "org.freedesktop.hostname1"
	dbusPath = "/org/freedesktop/hostname1"
)

// Conn is a connection to systemds dbus endpoint.
type Conn struct {
	conn   *dbus.Conn
	object dbus.BusObject
}

// New establishes a connection to the system bus and authenticates.
func New() (*Conn, error) {
	c := new(Conn)

	if err := c.initConnection(); err != nil {
		return nil, err
	}

	return c, nil
}

// Close closes the dbus connection
func (c *Conn) Close() {
	if c == nil {
		return
	}

	if c.conn != nil {
		c.conn.Close()
	}
}

// Connected returns whether conn is connected
func (c *Conn) Connected() bool {
	return c.conn.Connected()
}

func (c *Conn) initConnection() error {
	var err error
	if c.conn, err = dbus.SystemBusPrivate(); err != nil {
		return err
	}

	// Only use EXTERNAL method, and hardcode the uid (not username)
	// to avoid a username lookup (which requires a dynamically linked
	// libc)
	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}

	if err := c.conn.Auth(methods); err != nil {
		c.conn.Close()
		return err
	}

	if err := c.conn.Hello(); err != nil {
		c.conn.Close()
		return err
	}

	c.object = c.conn.Object(dbusDest, dbus.ObjectPath(dbusPath))

	return nil
}

func (c *Conn) SetHostname(hostname string, interactive bool) error {
	return c.object.Call(dbusDest+".SetHostname", 0, hostname, interactive).Err
}

func (c *Conn) SetStaticHostname(hostname string, interactive bool) error {
	return c.object.Call(dbusDest+".SetStaticHostname", 0, hostname, interactive).Err
}

func (c *Conn) SetPrettyHostname(hostname string, interactive bool) error {
	return c.object.Call(dbusDest+".SetPrettyHostname", 0, hostname, interactive).Err
}

func (c *Conn) SetIconName(iconName string, interactive bool) error {
	return c.object.Call(dbusDest+".SetIconName", 0, iconName, interactive).Err
}

func (c *Conn) SetChassis(chassis string, interactive bool) error {
	return c.object.Call(dbusDest+".SetChassis", 0, chassis, interactive).Err
}

func (c *Conn) SetDeployment(deployment string, interactive bool) error {
	return c.object.Call(dbusDest+".SetDeployment", 0, deployment, interactive).Err
}

func (c *Conn) SetLocation(location string, interactive bool) error {
	return c.object.Call(dbusDest+".SetLocation", 0, location, interactive).Err
}

func (c *Conn) GetProductUUID(interactive bool) ([]byte, error) {
	var uuid []byte
	err := c.object.Call(dbusDest+".GetProductUUID", 0, interactive).Store(&uuid)
	return uuid, err
}

func (c *Conn) GetHardwareSerial() (string, error) {
	var serial string
	err := c.object.Call(dbusDest+".GetHardwareSerial", 0).Store(&serial)
	return serial, err
}

func (c *Conn) Describe() (string, error) {
	var description string
	err := c.object.Call(dbusDest+".Describe", 0).Store(&description)
	return description, err
}

func (c *Conn) Hostname() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".Hostname")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) StaticHostname() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".StaticHostname")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) PrettyHostname() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".PrettyHostname")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) DefaultHostname() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".DefaultHostname")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) HostnameSource() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".HostnameSource")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) IconName() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".IconName")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) Chassis() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".Chassis")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) Deployment() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".Deployment")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) Location() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".Location")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) KernelName() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".KernelName")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) KernelRelease() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".KernelRelease")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) KernelVersion() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".KernelVersion")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) OperatingSystemPrettyName() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".OperatingSystemPrettyName")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) OperatingSystemCPEName() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".OperatingSystemCPEName")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) OperatingSystemSupportEnd() (uint64, error) {
	out, err := c.object.GetProperty(dbusDest + ".OperatingSystemSupportEnd")
	if err != nil {
		return 0, err
	}
	return out.Value().(uint64), nil
}

func (c *Conn) HomeURL() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".HomeURL")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) HardwareVendor() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".HardwareVendor")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) HardwareModel() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".HardwareModel")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) FirmwareVersion() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".FirmwareVersion")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) FirmwareVendor() (string, error) {
	out, err := c.object.GetProperty(dbusDest + ".FirmwareVendor")
	if err != nil {
		return "", err
	}
	return out.Value().(string), nil
}

func (c *Conn) FirmwareDate() (uint64, error) {
	out, err := c.object.GetProperty(dbusDest + ".FirmwareDate")
	if err != nil {
		return 0, err
	}
	return out.Value().(uint64), nil
}
