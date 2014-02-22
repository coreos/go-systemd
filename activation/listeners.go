/*
Copyright 2014 CoreOS Inc.

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

package activation

import (
	"fmt"
	"net"
	"os"
)

type TcpOrUdp struct {
	Tcp net.Listener
	Udp net.PacketConn
	Err error
}

// WrapSystemdSockets will take a list of files from Files function and create
// UDP sockets or TCP listeners for them.
func WrapSystemdSockets(files []*os.File) (result []TcpOrUdp) {
	result = make([]TcpOrUdp, len(files))
	for i, fd := range files {
		if pc, err := net.FilePacketConn(fd); err == nil {
			result[i].Udp = pc
			continue
		}
		if sc, err := net.FileListener(fd); err == nil {
			result[i].Tcp = sc
			continue
		} else {
			result[i].Err = err
		}
	}
	return
}

// Listeners returns net.Listeners for all socket activated fds passed to this process.
func Listeners(unsetEnv bool) ([]net.Listener, error) {
	files := Files(unsetEnv)
	listeners := make([]net.Listener, len(files))

	for i, f := range WrapSystemdSockets(files) {
		if f.Err != nil {
			return nil, fmt.Errorf("Error setting up FileListener for fd %d: %s", files[i].Fd(), f.Err.Error())
		}
		listeners[i] = f.Tcp
	}

	return listeners, nil
}
