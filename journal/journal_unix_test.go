// Copyright 2022 CoreOS, Inc.
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

//go:build !windows
// +build !windows

package journal_test

import (
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/coreos/go-systemd/v22/journal"
)

func TestStderrIsJournalStream(t *testing.T) {
	if _, ok := os.LookupEnv("JOURNAL_STREAM"); ok {
		t.Fatal("unset JOURNAL_STREAM before running this test")
	}

	t.Run("Missing", func(t *testing.T) {
		ok, err := journal.StderrIsJournalStream()
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Error("stderr shouldn't be connected to journal stream")
		}
	})
	t.Run("Present", func(t *testing.T) {
		f, stat := getUnixStreamSocket(t)
		defer f.Close()
		os.Setenv("JOURNAL_STREAM", fmt.Sprintf("%d:%d", stat.Dev, stat.Ino))
		defer os.Unsetenv("JOURNAL_STREAM")
		replaceStderr(int(f.Fd()), func() {
			ok, err := journal.StderrIsJournalStream()
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Error("stderr should've been connected to journal stream")
			}
		})
	})
	t.Run("NotMatching", func(t *testing.T) {
		f, stat := getUnixStreamSocket(t)
		defer f.Close()
		os.Setenv("JOURNAL_STREAM", fmt.Sprintf("%d:%d", stat.Dev+1, stat.Ino))
		defer os.Unsetenv("JOURNAL_STREAM")
		replaceStderr(int(f.Fd()), func() {
			ok, err := journal.StderrIsJournalStream()
			if err != nil {
				t.Fatal(err)
			}
			if ok {
				t.Error("stderr shouldn't be connected to journal stream")
			}
		})
	})
	t.Run("Malformed", func(t *testing.T) {
		f, stat := getUnixStreamSocket(t)
		defer f.Close()
		os.Setenv("JOURNAL_STREAM", fmt.Sprintf("%d-%d", stat.Dev, stat.Ino))
		defer os.Unsetenv("JOURNAL_STREAM")
		replaceStderr(int(f.Fd()), func() {
			_, err := journal.StderrIsJournalStream()
			if err == nil {
				t.Fatal("JOURNAL_STREAM is malformed, but no error returned")
			}
		})
	})
}

func ExampleStderrIsJournalStream() {
	// NOTE: this is just an example. Production code
	// will likely use this to setup a logging library
	// to write messages to either journal or stderr.
	ok, err := journal.StderrIsJournalStream()
	if err != nil {
		panic(err)
	}

	if ok {
		// use journal native protocol
		journal.Send("this is a message logged through the native protocol", journal.PriInfo, nil)
	} else {
		// use stderr
		fmt.Fprintln(os.Stderr, "this is a message logged through stderr")
	}
}

func replaceStderr(fd int, cb func()) {
	savedStderr, err := syscall.Dup(syscall.Stderr)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(savedStderr)
	err = syscall.Dup2(fd, syscall.Stderr)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := syscall.Dup2(savedStderr, syscall.Stderr)
		if err != nil {
			panic(err)
		}
	}()
	cb()
}

// getUnixStreamSocket returns a unix stream socket obtained with
// socketpair(2), and its fstat result. Only one end of the socket pair
// is returned, and the other end is closed immediately: we don't need
// it for our purposes.
func getUnixStreamSocket(t *testing.T) (*os.File, *syscall.Stat_t) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(os.NewSyscallError("socketpair", err))
	}
	// we don't need the remote end for our tests
	syscall.Close(fds[1])

	file := os.NewFile(uintptr(fds[0]), "unix-stream")
	stat, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	return file, stat.Sys().(*syscall.Stat_t)
}
