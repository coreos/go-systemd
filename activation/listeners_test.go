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

package activation

import (
	"net"
	"os"
	"os/exec"
	"testing"
)

// TestActivation forks out a copy of activation.go example and reads back two
// strings from the pipes that are passed in.
func TestListeners(t *testing.T) {
	arg0, cmdline := exampleCmd("listen")
	cmd := exec.Command(arg0, cmdline...)

	l1, err := net.Listen("tcp", ":9999")
	if err != nil {
		t.Fatal(err)
	}
	l2, err := net.Listen("tcp", ":1234")
	if err != nil {
		t.Fatal(err)
	}

	t1 := l1.(*net.TCPListener)
	t2 := l2.(*net.TCPListener)

	f1, _ := t1.File()
	f2, _ := t2.File()

	cmd.ExtraFiles = []*os.File{
		f1,
		f2,
	}

	r1, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r1.Write([]byte("Hi")); err != nil {
		t.Fatal(err)
	}

	r2, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r2.Write([]byte("Hi")); err != nil {
		t.Fatal(err)
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "LISTEN_FDS=2", "LISTEN_FDNAMES=fd1:fd2", "FIX_LISTEN_PID=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Unexpected error: %v (command output: %s)", err, out)
	}

	correctStringWritten(t, r1, "Hello world: fd1")
	correctStringWritten(t, r2, "Goodbye world: fd2")
}
