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

// Package util contains utility functions related to systemd that applications
// can use to check things like whether systemd is running.  Note that some of
// these functions attempt to manually load systemd libraries at runtime rather
// than linking against them.
package util

// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include <dlfcn.h>
// #include <sys/types.h>
//
// int
// my_sd_pid_get_owner_uid(void *f, pid_t pid, uid_t *uid)
// {
//   int (*sd_pid_get_owner_uid)(pid_t, uid_t *);
//
//   sd_pid_get_owner_uid = (int (*)(pid_t, uid_t *))f;
//   return sd_pid_get_owner_uid(pid, uid);
// }
//
// int
// my_sd_pid_get_slice(void *f, pid_t pid, char **slice)
// {
//   int (*sd_pid_get_slice)(pid_t, char **);
//
//   sd_pid_get_slice = (int (*)(pid_t, char **))f;
//   return sd_pid_get_slice(pid, slice);
// }
//
import "C"
import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// getHandle tries to get a handle to a systemd library (.so), attempting to
// access it by several different names and returning the first that is
// successfully opened. Callers are responsible for closing the handler.
// Returns nil if no library can be found.
func getHandle() (unsafe.Pointer, string) {
	for _, name := range []string{
		// systemd < 209
		"libsystemd-login.so",
		"libsystemd-login.so.0",

		// systemd >= 209 merged libsystemd-login into libsystemd proper
		"libsystemd.so",
		"libsystemd.so.0",
	} {
		libname := C.CString(name)
		defer C.free(unsafe.Pointer(libname))
		handle := C.dlopen(libname, C.RTLD_LAZY)
		if handle != nil {
			return handle, name
		}
	}
	return nil, ""
}

// GetRunningSlice attempts to retrieve the name of the systemd slice in which
// the current process is running.
func GetRunningSlice() (slice string, err error) {
	handle, libname := getHandle()
	if handle == nil {
		err = fmt.Errorf("error opening libsystemd-login.so")
		return
	}
	defer func() {
		if r := C.dlclose(handle); r != 0 {
			err = fmt.Errorf("error closing %v", libname)
		}
	}()

	sym := C.CString("sd_pid_get_slice")
	defer C.free(unsafe.Pointer(sym))
	sd_pid_get_slice := C.dlsym(handle, sym)
	if sd_pid_get_slice == nil {
		err = fmt.Errorf("error resolving sd_pid_get_slice function")
		return
	}

	var s string
	sl := C.CString(s)
	defer C.free(unsafe.Pointer(sl))

	ret := C.my_sd_pid_get_slice(sd_pid_get_slice, 0, &sl)
	if ret < 0 {
		err = fmt.Errorf("error calling sd_pid_get_slice: %v", syscall.Errno(-ret))
		return
	}

	slice = C.GoString(sl)
	return
}

// RunningFromSystemService detects whether the current process has been invoked
// from a system service. The condition for this is whether the process is
// _not_ a user process. User processes are those running in session scopes or
// under per-user `systemd --user` instances
func RunningFromSystemService() (ret bool, err error) {
	handle, libname := getHandle()
	if handle == nil {
		// can't open libsystemd so we assume systemd is not
		// installed and we're not running from a unit file
		ret = false
		return
	}
	defer func() {
		if r := C.dlclose(handle); r != 0 {
			err = fmt.Errorf("error closing %v", libname)
		}
	}()

	sym := C.CString("sd_pid_get_owner_uid")
	defer C.free(unsafe.Pointer(sym))
	sd_pid_get_owner_uid := C.dlsym(handle, sym)
	if sd_pid_get_owner_uid == nil {
		err = fmt.Errorf("error resolving sd_pid_get_owner_uid function")
		return
	}

	var uid C.uid_t
	errno := C.my_sd_pid_get_owner_uid(sd_pid_get_owner_uid, 0, &uid)
	serrno := syscall.Errno(-errno)
	// when we're running from a unit file, sd_pid_get_owner_uid returns
	// ENOENT (systemd <220) or ENXIO (systemd >=220)
	switch {
	case errno >= 0:
		ret = false
		return
	case serrno == syscall.ENOENT, serrno == syscall.ENXIO:
		ret = true
		return
	default:
		err = fmt.Errorf("error calling sd_pid_get_owner_uid: %v", syscall.Errno(-errno))
		return
	}
}

// IsRunningSystemd checks whether the host was booted with systemd as its init
// system. This functions similar to systemd's `sd_booted(3)`: internally, it
// checks whether /run/systemd/system/ exists and is a directory.
// http://www.freedesktop.org/software/systemd/man/sd_booted.html
func IsRunningSystemd() bool {
	fi, err := os.Lstat("/run/systemd/system")
	if err != nil {
		return false
	}
	return fi.IsDir()
}
