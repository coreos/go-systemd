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
	"fmt"
	"os/exec"
	"testing"
)

var (
	conn        *Conn
	machineName = "testMachine"
)

func createTestProcess() (pid int, err error) {
	systemdRun, lookErr := exec.LookPath("systemd-run")
	if lookErr != nil {
		return -1, lookErr
	}
	sleep, lookErr := exec.LookPath("sleep")
	if lookErr != nil {
		return -1, lookErr
	}
	cmd := exec.Command(systemdRun, sleep, "5000")
	err = cmd.Run()
	if err != nil {
		return -1, err
	}
	return cmd.Process.Pid + 1, nil
}

// TestNew ensures that New() works without errors.
func TestNew(t *testing.T) {
	c, err := New()
	conn = c
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegister(t *testing.T) {
	leader, lErr := createTestProcess()
	if lErr != nil {
		t.Error(lErr)
	}
	t.Log(leader)
	regErr := conn.RegisterMachine(machineName, nil, "go-systemd", "container", leader, "")
	if regErr != nil {
		t.Error(regErr)
	}
}

func TestGet(t *testing.T) {
	machine, machineErr := conn.GetMachine(machineName)
	if machineErr != nil {
		t.Error(machineErr)
	}
	if len(machine) == 0 {
		t.Error(fmt.Errorf("did not find machine named %s", machineName))
	}
}

func TestTerminate(t *testing.T) {
	tErr := conn.TerminateMachine(machineName)
	if tErr != nil {
		t.Error(tErr)
	}
}
