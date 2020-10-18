// Copyright 2020 CoreOS, Inc.
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

package daemon

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestSdNotifyBarrier(t *testing.T) {

	testDir, e := ioutil.TempDir("/tmp/", "test-")
	if e != nil {
		panic(e)
	}
	defer os.RemoveAll(testDir)

	notifySocket := testDir + "/notify-socket.sock"
	laddr := net.UnixAddr{
		Name: notifySocket,
		Net:  "unixgram",
	}
	sock, err := net.ListenUnixgram("unixgram", &laddr)
	if err != nil {
		panic(err)
	}

	messageExpected := []byte("BARRIER=1")

	tests := []struct {
		unsetEnv       bool
		envSocket      string
		expectErr      string
		expectReadN    int // num in-band bytes to recv on socket
		expectReadOobN int // num out-of-band bytes to recv on socket
	}{
		// should succeed
		{
			unsetEnv:       false,
			envSocket:      notifySocket,
			expectErr:      "",
			expectReadN:    len(messageExpected),
			expectReadOobN: syscall.CmsgSpace(4 /*1xFD*/),
		},
		// failure to open systemd socket should result in an error
		{
			unsetEnv:       false,
			envSocket:      testDir + "/missing.sock",
			expectErr:      "no such file",
			expectReadN:    0,
			expectReadOobN: 0,
		},
		// notification not supported
		{
			unsetEnv:       false,
			envSocket:      "",
			expectErr:      ErrEnvironment.Error(),
			expectReadN:    0,
			expectReadOobN: 0,
		},
	}

	resultCh := make(chan error)

	// allocate message and out-of-band buffers
	var msgBuf [128]byte
	oobBuf := make([]byte, syscall.CmsgSpace(4 /*1xFD/1xint32*/))

	for i, tt := range tests {
		must(os.Unsetenv("NOTIFY_SOCKET"))
		if tt.envSocket != "" {
			must(os.Setenv("NOTIFY_SOCKET", tt.envSocket))
		}

		go func() {
			ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
			resultCh <- SdNotifyBarrier(ctx, tt.unsetEnv)
		}()

		if tt.envSocket == notifySocket {
			// pretend to be systemd and read the message that SdNotifyBarrier wrote to sock
			// returns (n, oobn, flags int, addr *UnixAddr, err error)
			n, oobn, _, _, err := sock.ReadMsgUnix(msgBuf[:], oobBuf[:])
			// fmt.Printf("ReadMsgUnix -> %v, %v, %v, %v, %v\n", n, oobn, flags, from, err)
			if err != nil {
				t.Errorf("#%d: failed to read socket: %v", i, err)
				continue
			}

			// check bytes read
			if tt.expectReadN != n {
				t.Errorf("#%d: want expectReadN %v, got %v", i, tt.expectReadN, n)
				continue
			}
			if tt.expectReadOobN != oobn {
				t.Errorf("#%d: want expectReadOobN %v, got %v", i, tt.expectReadOobN, n)
				continue
			}

			// check message
			if n > 0 {
				if !bytes.Equal(msgBuf[:n], messageExpected) {
					t.Errorf("#%d: want message %q, got %q", i, messageExpected, msgBuf[:n])
					continue
				}
			}

			// parse OOB message
			if oobn > 0 {
				mv, err := syscall.ParseSocketControlMessage(oobBuf)
				if err != nil {
					t.Errorf("#%d: ParseSocketControlMessage failed: %v", i, err)
					continue
				}

				if len(mv) != 1 {
					// should be just control message in the oob data
					t.Errorf("#%d: want len(mv)=1, got %v", i, len(mv))
					continue
				}

				// parse socket fd from message 0
				fds, err := syscall.ParseUnixRights(&mv[0])
				if err != nil {
					t.Errorf("#%d: ParseUnixRights failed: %v", i, err)
					continue
				}
				if len(fds) != 1 {
					// should be just one socket file descriptor in the control message
					t.Errorf("#%d: want len(fds)=1, got %v", i, len(fds))
					continue
				}

				// finally close the socket to signal back to SdNotifyBarrier
				syscall.Close(fds[0])
			}
		} // if tt.envSocket == notifySocket

		err = <-resultCh

		// check error
		if len(tt.expectErr) > 0 {
			if err == nil {
				t.Errorf("#%d: want non-nil err, got nil", i)
			} else if !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("#%d: want err with substr %q, got %q", i, tt.expectErr, err.Error())
			}
		} else if len(tt.expectErr) == 0 && err != nil {
			t.Errorf("#%d: want nil err, got %v", i, err)
		}

		// if unsetEnvironment was requested, verify NOTIFY_SOCKET is not set
		if tt.unsetEnv && tt.envSocket != "" && os.Getenv("NOTIFY_SOCKET") != "" {
			t.Errorf("#%d: environment variable not cleaned up", i)
		}

	}
}
