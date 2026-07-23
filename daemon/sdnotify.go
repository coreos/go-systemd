// Copyright 2014 Docker, Inc.
// Copyright 2015-2018 CoreOS, Inc.
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
//

// Package daemon provides a Go implementation of the sd_notify protocol.
// It can be used to inform systemd of service start-up completion, watchdog
// events, and other status changes.
//
// https://www.freedesktop.org/software/systemd/man/sd_notify.html#Description
package daemon

import (
	"context"
	"io"
	"net"
	"os"
	"time"
)

const (
	// SdNotifyReady tells the service manager that service startup is finished
	// or the service finished loading its configuration.
	SdNotifyReady = "READY=1"

	// SdNotifyStopping tells the service manager that the service is beginning
	// its shutdown.
	SdNotifyStopping = "STOPPING=1"

	// SdNotifyReloading tells the service manager that this service is
	// reloading its configuration. Note that you must call SdNotifyReady when
	// it completed reloading.
	SdNotifyReloading = "RELOADING=1"

	// SdNotifyWatchdog tells the service manager to update the watchdog
	// timestamp for the service.
	SdNotifyWatchdog = "WATCHDOG=1"
)

var aLongTimeAgo = time.Unix(1, 0)

// SdNotify sends a message to the init daemon. It is common to ignore the error.
// If `unsetEnvironment` is true, the environment variable `NOTIFY_SOCKET`
// will be unconditionally unset.
//
// It returns one of the following:
// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
// (true, nil) - notification supported, data has been sent
func SdNotify(unsetEnvironment bool, state string) (bool, error) {
	return SdNotifyWithFiles(unsetEnvironment, state)
}

// SdNotifyWithFiles is like [SdNotify], but additionally passes the given open
// files to the service manager. It is typically used together with "FDSTORE=1"
// to hand file descriptors over for safe-keeping across a restart.
//
// The files remain owned by the caller; they may be closed as soon as this
// function returns, if necessary.
//
// The return values are the same as for [SdNotify].
func SdNotifyWithFiles(unsetEnvironment bool, state string, files ...*os.File) (bool, error) {
	socketAddr := &net.UnixAddr{
		Name: os.Getenv("NOTIFY_SOCKET"),
		Net:  "unixgram",
	}

	// NOTIFY_SOCKET not set
	if socketAddr.Name == "" {
		return false, nil
	}

	if unsetEnvironment {
		if err := os.Unsetenv("NOTIFY_SOCKET"); err != nil {
			return false, err
		}
	}

	// Use an unconnected socket instead of [net.DialUnix] because
	// [net.UnixConn.WriteMsgUnix] needs datagram sockets to be
	// unconnected.
	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{})
	if err != nil {
		return false, err
	}
	defer conn.Close()

	if err := writeState(conn, socketAddr, state, files); err != nil {
		return false, err
	}
	return true, nil
}

// SdNotifyBarrier blocks until all notifications sent before it have been
// processed by the service manager. This is the equivalent of libsystemd's
// sd_notify_barrier(3).
//
// It works by sending a "BARRIER=1" message together with the write end of a
// freshly created pipe. The service manager closes its copy of that descriptor
// once it has handled every preceding notification, at which point our read end
// observes EOF and the barrier is lifted.
//
// It returns one of the following:
// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET, while sending data)
// (true, err) - notification supported and sent, but barrier hasn't been reached
// (true, nil) - notification supported and sent, and the barrier has been reached
func SdNotifyBarrier(ctx context.Context, unsetEnvironment bool) (bool, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return false, err
	}
	defer r.Close()

	// Send the barrier along with the write end of the pipe. We close our copy
	// of the write end immediately afterwards (whatever the outcome), so that
	// the service manager holds the only remaining reference to it.
	sent, err := SdNotifyWithFiles(unsetEnvironment, "BARRIER=1", w)
	w.Close()
	if err != nil || !sent {
		return sent, err
	}

	err = waitForClose(ctx, r)
	return true, err
}

func waitForClose(ctx context.Context, r *os.File) error {
	// interrupt read on context cancelation
	stop := context.AfterFunc(ctx, func() {
		// error should never happen: deadlines are supported for pipes
		_ = r.SetReadDeadline(aLongTimeAgo)
	})
	defer stop()

	_, err := io.Copy(io.Discard, r)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}
	return nil
}
