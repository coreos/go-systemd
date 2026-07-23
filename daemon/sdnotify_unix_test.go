// Copyright 2026
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

//go:build unix

package daemon

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func recvNotifyFiles(t *testing.T, conn *net.UnixConn) (string, []*os.File) {
	t.Helper()

	buf := make([]byte, 4096)
	oob := make([]byte, unix.CmsgSpace(1024))
	n, oobn, _, _, err := conn.ReadMsgUnix(buf, oob)
	if err != nil {
		t.Fatalf("ReadMsgUnix: %v", err)
	}

	var files []*os.File
	scms, err := unix.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		t.Fatalf("ParseSocketControlMessage: %v", err)
	}
	for _, scm := range scms {
		fds, err := unix.ParseUnixRights(&scm)
		if err != nil {
			t.Fatalf("ParseUnixRights: %v", err)
		}
		for _, fd := range fds {
			files = append(files, os.NewFile(uintptr(fd), ""))
		}
	}
	return string(buf[:n]), files
}

func TestSdNotifyWithFiles(t *testing.T) {
	const numStoredFiles = 3

	testDir := t.TempDir()

	notifySocket := testDir + "/notify-socket.sock"
	laddr := net.UnixAddr{
		Name: notifySocket,
		Net:  "unixgram",
	}
	conn, err := net.ListenUnixgram("unixgram", &laddr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var storedFiles []*os.File

	for range numStoredFiles {
		f, err := os.CreateTemp(testDir, "")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		storedFiles = append(storedFiles, f)
	}

	t.Setenv("NOTIFY_SOCKET", notifySocket)

	// Store file descriptors in the test service manager

	state := "FDSTORE=1"

	sent, err := SdNotifyWithFiles(false, "FDSTORE=1", storedFiles...)
	if err != nil {
		t.Fatalf("SdNotifyWithFiles returned error: %v", err)
	}
	if !sent {
		t.Fatal("SdNotifyWithFiles reported the notification was not sent")
	}

	// Note that we don't close the original files, we don't have to.

	// And we're now the "service manager".

	receivedState, fds := recvNotifyFiles(t, conn)

	if receivedState != state {
		t.Errorf("received state = %q; want %q", receivedState, state)
	}

	if len(fds) != len(storedFiles) {
		t.Fatalf("received %d file descriptors; want %d", len(fds), len(storedFiles))
	}

	// The received file desctiptors must refer to the same files
	for i, f := range fds {
		defer f.Close()

		wantStat, err := storedFiles[i].Stat()
		if err != nil {
			t.Fatal(err)
		}

		gotStat, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}

		if !os.SameFile(gotStat, wantStat) {
			// "%#v" for os.fileStat is ugly, but decipherable
			t.Errorf("file %d: wrong file descriptor received; got %#v, want %#v", i, gotStat, wantStat)
		}
	}
}

func TestSdNotifyBarrier(t *testing.T) {
	tests := []struct {
		name    string
		timeout bool
	}{
		{name: "OK", timeout: false},
		{name: "Timeout", timeout: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := t.TempDir()

			notifySocket := testDir + "/notify-socket.sock"
			laddr := net.UnixAddr{
				Name: notifySocket,
				Net:  "unixgram",
			}
			conn, err := net.ListenUnixgram("unixgram", &laddr)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			t.Setenv("NOTIFY_SOCKET", notifySocket)

			ctx, cancel := context.WithCancel(t.Context())

			var t1, t2 time.Time

			go func() {
				_, files := recvNotifyFiles(t, conn)
				if tt.timeout {
					// simulate timeout for the SdNotifyBarrier call
					cancel()
					<-t.Context().Done()
				} else {
					t1 = time.Now()
					files[0].Close()
				}
			}()

			sent, err := SdNotifyBarrier(ctx, false)
			if !sent {
				t.Fatal("SdNotifyBarrier should've sent a message")
			}
			t2 = time.Now()

			if tt.timeout {
				if !errors.Is(err, context.Canceled) {
					t.Errorf("unexpected error %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
				if t1.After(t2) {
					t.Error("SdNotifyWithFiles returned before service manager closed the file")
				}
			}
		})
	}
}
