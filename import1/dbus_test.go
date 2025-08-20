// Copyright 2019 CoreOS, Inc.
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

package import1

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	importPrefix = "importd-test-"
)

func cleanupImport(t *testing.T, name string) {
	t.Cleanup(func() {
		var (
			out []byte
			err error
			dur = 500 * time.Millisecond
		)
		for range 5 {
			time.Sleep(dur)
			out, err = exec.Command("machinectl", "remove", name).CombinedOutput()
			if err == nil {
				return
			}
			dur *= 2
		}
		t.Fatalf("machinectl remove %s failed: %v\noutput: %s", name, err, out)
	})
}

func TestImportTar(t *testing.T) {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}

	f := openFixture(t, "image.tar.xz")
	name := importPrefix + "ImportTar"
	cleanupImport(t, name)
	_, err = conn.ImportTar(f, name, true, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestImportRaw(t *testing.T) {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}

	f := openFixture(t, "image.raw.xz")
	name := importPrefix + "ImportRaw"
	cleanupImport(t, name)
	_, err = conn.ImportRaw(f, name, true, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExportTar(t *testing.T) {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(t.TempDir() + "image-export.tar.xz")
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.ExportTar(importPrefix+"ImportTar", f, "xz")
	if err != nil {
		t.Fatal(err)
	}
}

func TestExportRaw(t *testing.T) {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(t.TempDir() + "image-export.raw.xz")
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.ExportRaw(importPrefix+"ImportRaw", f, "xz")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPullTar(t *testing.T) {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.PullTar("http://127.0.0.1:8080/image.tar.xz", importPrefix+"PullTar", "no", true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPullRaw(t *testing.T) {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.PullRaw("http://127.0.0.1:8080/image.raw.xz", importPrefix+"PullRaw", "no", true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestListAndCancelTransfers(t *testing.T) {
	conn, err := New()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		_, _ = conn.PullTar("http://127.0.0.1:8080/image.tar.xz", importPrefix+"ListAndCancelTransfers", "no", true)
		_, _ = conn.PullTar("http://127.0.0.1:8080/image.raw.xz", importPrefix+"ListAndCancelTransfers", "no", true)
	}()

	transfers, err := conn.ListTransfers()
	if err != nil {
		t.Error(err)
	}
	if len(transfers) < 1 {
		t.Error("transfers length is not correct")
	}

	for _, v := range transfers {
		err = conn.CancelTransfer(v.Id)
		if err != nil {
			// Let's just ignore the transfer id not found error.
			if strings.Contains(err.Error(), "No transfer by id") {
				continue
			}
			t.Error(err)
		}
	}
}

func openFixture(t *testing.T, target string) *os.File {
	abs, err := filepath.Abs("../fixtures/" + target)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(abs)
	if err != nil {
		t.Fatal(err)
	}

	return f
}

func init() {
	go func() {
		err := http.ListenAndServe(":8080", http.FileServer(http.Dir("../fixtures")))
		if err != nil {
			log.Fatal(err)
		}
	}()
}
