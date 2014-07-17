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
	"net"
)

// PacketConns returns net.PacketConns for all applicable socket activated fds
// passed to this process.
func PacketConns(unsetEnv bool) ([]net.PacketConn, error) {
	files := Files(unsetEnv)
	conns := make([]net.PacketConn, 0)
	for _, f := range files {
		if pc, err := net.FilePacketConn(f); err == nil {
			conns = append(conns, pc)
			continue
		}
	}
	return conns, nil
}
