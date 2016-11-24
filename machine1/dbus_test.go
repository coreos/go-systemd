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
	"os"
	"testing"
)

// TestNew ensures that New() works without errors.
func TestNew(t *testing.T) {
	_, err := New()

	if err != nil {
		t.Fatal(err)
	}
}

var machineName = "testMachine"

func TestRegister(t *testing.T) {
	conn, connErr := New()
	if connErr != nil {
		t.Error(connErr)
	}
	regErr := conn.RegisterMachine(machineName, []byte{}, "go-systemd", "container", os.Getpid(), "")
	if regErr != nil {
		t.Error(regErr)
	}
}

func TestGet(t *testing.T) {
	conn, connErr := New()
	if connErr != nil {
		t.Error(connErr)
	}
	machine, machineErr := conn.GetMachine(machineName)
	if machineErr != nil {
		t.Error(machineErr)
	}
	if len(machine) == 0 {
		t.Error(fmt.Errorf("did not find machine named %s", machineName))
	}
}

func TestTerminate(t *testing.T) {
	conn, connErr := New()
	if connErr != nil {
		t.Error(connErr)
	}
	tErr := conn.TerminateMachine(machineName)
	if tErr != nil {
		t.Error(tErr)
	}
}
