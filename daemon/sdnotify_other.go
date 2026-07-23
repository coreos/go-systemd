// Copyright 2025
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !unix

package daemon

import (
	"errors"
	"net"
	"os"
)

// SdNotifyMonotonicUsec returns the empty string on unsupported platforms.
func SdNotifyMonotonicUsec() string {
	return ""
}

// writeState sends the notification state. Passing files is not supported on
// this platform, so a non-empty files slice results in an error.
func writeState(conn *net.UnixConn, addr *net.UnixAddr, state string, files []*os.File) error {
	if len(files) > 0 {
		return errors.New("daemon: passing files over NOTIFY_SOCKET is not supported on this platform")
	}
	_, err := conn.WriteToUnix([]byte(state), addr)
	return err
}
