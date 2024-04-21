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

//go:build !windows
// +build !windows

package activation

import (
	"fmt"
	"os"
	"syscall"
)

func (m Method) Apply(f *os.File) error {
	saveFd := int(f.Fd()) // get the idx before being closed.

	switch m {
	case ConsumeFiles:
		f.Close()
	case ReserveFiles:
		devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0755)
		if err != nil {
			return fmt.Errorf("accessing /dev/null: %w", err)
		}

		nullFd := int(devNull.Fd())

		// "If oldfd equals newfd, then dup3() fails with the error EINVAL."
		if saveFd == nullFd {
			syscall.CloseOnExec(nullFd)
		} else {
			// "makes newfd be the copy of oldfd, closing newfd first if necessary"
			if err := syscall.Dup3(nullFd, saveFd, syscall.O_CLOEXEC); err != nil {
				return fmt.Errorf("setting %d fd to /dev/null: %w", saveFd, err)
			}
		}
	case ConserveFiles:
		// no action
	}

	return nil
}
