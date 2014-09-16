/*
Copyright 2014 CoreOS Inc.

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

package dbus

import (
	"github.com/godbus/dbus"
)

const (
	signalBuffer = 1000
)

func (c *Conn) initDispatch() {
	// Setup the listeners on jobs so that we can get completions
	c.sysconn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal', interface='org.freedesktop.systemd1.Manager', member='JobRemoved'")

	sigCh := make(chan *dbus.Signal, signalBuffer)
	c.sysconn.Signal(sigCh)

	go c.dispatch(sigCh)
}

type jobRequest struct {
	name     string
	args     []interface{}
	resultCh chan *jobResult
}

type jobResult struct {
	output string
	err    error
}

func (c *Conn) dispatch(sigCh chan *dbus.Signal) {
	jMap := make(map[dbus.ObjectPath]chan *jobResult)

	// these are needed to call dbus.Store, but the values don't matter
	var _uint32 uint32
	var _string string

	handleRequest := func(req *jobRequest) {
		var path dbus.ObjectPath
		err := c.sysobj.Call(req.name, 0, req.args...).Store(&path)
		if err != nil {
			req.resultCh <- &jobResult{err: err}
		} else {
			jMap[path] = req.resultCh
		}
	}

	handleSignal := func(sig *dbus.Signal) {
		if sig.Name == "org.freedesktop.systemd1.Manager.JobRemoved" {
			var path dbus.ObjectPath
			var result string
			dbus.Store(sig.Body, &_uint32, &path, &_string, &result)

			out, ok := jMap[path]
			if ok {
				out <- &jobResult{output: result}
				delete(jMap, path)
			}
		}

		c.maybeSendSubStateUpdate(sig)
	}

	for {
		select {
		case req := <-c.jobReqCh:
			handleRequest(req)
		case sig := <-sigCh:
			handleSignal(sig)
		}
	}
}

func (c *Conn) maybeSendSubStateUpdate(sig *dbus.Signal) {
	if c.subscriber.updateCh == nil {
		return
	}

	var unitPath dbus.ObjectPath
	switch sig.Name {
	case "org.freedesktop.systemd1.Manager.JobRemoved":
		unitName := sig.Body[2].(string)
		c.sysobj.Call("org.freedesktop.systemd1.Manager.GetUnit", 0, unitName).Store(&unitPath)
	case "org.freedesktop.systemd1.Manager.UnitNew":
		unitPath = sig.Body[1].(dbus.ObjectPath)
	case "org.freedesktop.DBus.Properties.PropertiesChanged":
		if sig.Body[0].(string) == "org.freedesktop.systemd1.Unit" {
			unitPath = sig.Path
		}
	}

	if unitPath == dbus.ObjectPath("") {
		return
	}

	c.sendSubStateUpdate(unitPath)
}

func (c *Conn) runJob(name string, args ...interface{}) (string, error) {
	j := jobRequest{
		name:     name,
		args:     args,
		resultCh: make(chan *jobResult, 1),
	}

	c.jobReqCh <- &j

	res := <-j.resultCh
	return res.output, res.err
}
