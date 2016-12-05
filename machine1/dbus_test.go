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

package machine1

import (
	"os/exec"
	"testing"

	"github.com/coreos/go-systemd/dbus"
)

const (
	testServiceName = "machined-test.service"
	machineName     = "testMachine"
)

func mustCreateTestProcess() (pid int) {
	systemdRun, err := exec.LookPath("systemd-run")
	if err != nil {
		panic(err.Error())
	}
	sleep, err := exec.LookPath("sleep")
	if err != nil {
		panic(err.Error())
	}
	cmd := exec.Command(systemdRun, "--unit="+testServiceName, sleep, "5000")
	err = cmd.Run()
	if err != nil {
		panic(err.Error())
	}
	dbusConn, err := dbus.New()
	if err != nil {
		panic(err.Error())
	}
	defer dbusConn.Close()
	mainPIDProperty, err := dbusConn.GetServiceProperty(testServiceName, "MainPID")
	if err != nil {
		panic(err.Error())
	}
	mainPID := mainPIDProperty.Value.Value().(uint32)
	return int(mainPID)
}

func TestMachine(t *testing.T) {
	leader := mustCreateTestProcess()

	conn, newErr := New()
	if newErr != nil {
		t.Fatal(newErr)
	}

	regErr := conn.RegisterMachine(machineName, nil, "go-systemd", "container", leader, "")
	if regErr != nil {
		t.Fatal(regErr)
	}

	machine, getErr := conn.GetMachine(machineName)
	if getErr != nil {
		t.Fatal(getErr)
	}
	if len(machine) == 0 {
		t.Fatalf("did not find machine named %s", machineName)
	}

	tErr := conn.TerminateMachine(machineName)
	if tErr != nil {
		t.Fatal(tErr)
	}

	machine, getErr = conn.GetMachine(machineName)
	if len(machine) != 0 {
		t.Fatalf("unexpectedly found machine named %s", machineName)
	} else if getErr == nil {
		t.Fatal("expected error but got nil")
	}
}
