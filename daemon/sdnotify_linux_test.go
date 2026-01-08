// Copyright 2025
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
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// TestSdNotifyMonotonicUsec checks that SdNotifyMonotonicUsec is probably not returning complete garbage.
func TestSdNotifyMonotonicUsec(t *testing.T) {
	var resolution unix.Timespec
	if err := unix.ClockGetres(unix.CLOCK_MONOTONIC, &resolution); err != nil {
		if err == unix.EINVAL {
			t.Log("CLOCK_MONOTONIC is not supported on this system")
			if got := SdNotifyMonotonicUsec(); got != "" {
				t.Errorf("SdNotifyMonotonicUsec() = %q; want empty string", got)
			}
			return
		}
		t.Fatalf("ClockGetres(CLOCK_MONOTONIC) failed: %v", err)
	}

	now := func() uint64 {
		got := SdNotifyMonotonicUsec()
		t.Logf("SdNotifyMonotonicUsec() = %q", got)
		if got == "" {
			t.Fatal("SdNotifyMonotonicUsec() returned empty string on system which supports CLOCK_MONOTONIC")
		}
		fields := strings.SplitN(got, "=", 2)
		if len(fields) != 2 {
			t.Fatal("string is not a well-formed variable assignment")
		}
		tag, val := fields[0], fields[1]
		if tag != "MONOTONIC_USEC" {
			t.Errorf("expected tag MONOTONIC_USEC, got %q", tag)
		}
		if val[len(val)-1] != '\n' {
			t.Errorf("expected value to end with newline, got %q", val)
		}
		ts, err := strconv.ParseUint(val[:len(val)-1], 10, 64)
		if err != nil {
			t.Fatalf("value %q is not well-formed: %v", val, err)
		}
		if ts == 0 {
			// CLOCK_MONOTONIC is defined on Linux as the number of seconds
			// since boot, per clock_gettime(2). A timestamp of zero is
			// almost certainly bogus.
			t.Fatal("timestamp is zero")
		}
		return ts
	}

	start := now()
	time.Sleep(time.Duration(resolution.Nano()) * 3)
	ts := now()
	if ts < start {
		t.Errorf("timestamp went backwards: %d < %d", ts, start)
	} else if ts == start {
		t.Errorf("timestamp did not advance: %d == %d", ts, start)
	}
}
