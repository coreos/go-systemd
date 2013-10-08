// Integration with the systemd D-Bus API.  See http://www.freedesktop.org/wiki/Software/systemd/dbus/
package dbus

import (
	"github.com/guelfey/go.dbus"
	"sync"
)

const signalBuffer = 100

type Conn struct {
	sysconn     *dbus.Conn
	sysobj      *dbus.Object
	jobListener struct {
		jobs map[dbus.ObjectPath]chan string
		sync.Mutex
	}
}

func New() *Conn {
	c := new(Conn)
	c.initConnection()
	c.initJobs()
	c.initDispatch()
	return c
}

func (c *Conn) initConnection() {
	var err error
	c.sysconn, err = dbus.SystemBusPrivate()
	if err != nil {
		return
	}

	err = c.sysconn.Auth(nil)
	if err != nil {
		c.sysconn.Close()
		return
	}

	err = c.sysconn.Hello()
	if err != nil {
		c.sysconn.Close()
		return
	}

	c.sysobj = c.sysconn.Object("org.freedesktop.systemd1", dbus.ObjectPath("/org/freedesktop/systemd1"))

	c.sysconn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',interface='org.freedesktop.systemd1.Manager',member='JobRemoved'")

	err = c.sysobj.Call("org.freedesktop.systemd1.Manager.Subscribe", 0).Store()
	if err != nil {
		c.sysconn.Close()
		return
	}
}

func (c *Conn) initDispatch() {
	ch := make(chan *dbus.Signal, signalBuffer)

	c.sysconn.Signal(ch)

	go func() {
		for {
			signal := <-ch
			switch signal.Name {
			case "org.freedesktop.systemd1.Manager.JobRemoved":
				c.jobComplete(signal)
			}
		}
	}()
}
