// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package hostname1

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestNew ensures that New() works without errors.
func TestNew(t *testing.T) {
	if _, err := New(); err != nil {
		t.Fatal(err)
	}
}

// TestHostname ensures that the Hostname() method returns the system hostname.
func TestHostname(t *testing.T) {
	expectedHostname, err := exec.Command("hostname").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	h, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if hostname, err := h.Hostname(); err != nil {
		t.Fatal(err)
	} else if hostname != strings.TrimSuffix(string(expectedHostname), "\n") {
		t.Fatalf("expected %q, got %q", expectedHostname, hostname)
	}
}

func TestStaticHostname(t *testing.T) {
	hostnameFile, err := os.Open("/etc/hostname")
	if err != nil {
		t.Fatal(err)
	}
	defer hostnameFile.Close()

	expectedHostnameBytes := make([]byte, 256)
	n, err := hostnameFile.Read(expectedHostnameBytes)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	// Close the file so that hostnamed can use it.
	hostnameFile.Close()
	expectedHostname := strings.TrimSuffix(string(expectedHostnameBytes[:n]), "\n")

	h, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if hostname, err := h.StaticHostname(); err != nil {
		t.Fatal(err)
	} else if hostname != expectedHostname {
		t.Fatalf("expected %q, got %q", expectedHostname, hostname)
	}
}
