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
	"os"
	"syscall"
)

type ErrorDevNullSetup struct {
	fd  int
	err error
}

func (e ErrorDevNullSetup) Error() string {
	return "setting up %d fd to /dev/null: " + e.err.Error()
}

func (e ErrorDevNullSetup) Unwrap() error {
	return e.err
}

func (m Method) Apply(f *os.File) error {
	saveFd := int(f.Fd()) // get the idx before being closed.

	switch m {
	case ConsumeFiles:
		return f.Close()
	case ReserveFiles:
		devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0755)
		if err != nil {
			return ErrorDevNullSetup{err: err, fd: saveFd}
		}

		nullFd := int(devNull.Fd())

		// "If oldfd equals newfd, then dup3() fails with the error EINVAL."
		if saveFd == nullFd {
			syscall.CloseOnExec(nullFd)
		} else {
			// "makes newfd be the copy of oldfd, closing newfd first if necessary"
			if err := syscall.Dup3(nullFd, saveFd, syscall.O_CLOEXEC); err != nil {
				devNull.Close() // on an error tidy up.

				return ErrorDevNullSetup{err: err, fd: saveFd}
			}
		}
	case ConserveFiles:
		// no action
	}

	return nil
}
