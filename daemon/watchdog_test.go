// Copyright 2016 CoreOS, Inc.
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
	"os"
	"strconv"
	"testing"
)

func TestSdWatchdogEnabled(t *testing.T) {
	// (time, nil)
	err := os.Setenv("WATCHDOG_USEC", "100")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("WATCHDOG_PID", strconv.Itoa(os.Getpid()))
	if err != nil {
		panic(err)
	}

	delay, err := SdWatchdogEnabled()
	if delay == 0 || err != nil {
		t.Errorf("TEST: Watchdog enabled FAILED")
	}

	// (0, nil) PID doesnt match
	err = os.Setenv("WATCHDOG_USEC", "100")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("WATCHDOG_PID", "0")
	if err != nil {
		panic(err)
	}
	delay, err = SdWatchdogEnabled()
	if delay != 0 || err != nil {
		t.Errorf("TEST: PID doesn't match FAILED")
	}

	// (0, nil) WATCHDOG_USEC doen't exist
	err = os.Unsetenv("WATCHDOG_USEC")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("WATCHDOG_PID", strconv.Itoa(os.Getpid()))
	if err != nil {
		panic(err)
	}
	delay, err = SdWatchdogEnabled()
	if delay != 0 || err != nil {
		t.Errorf("TEST: WATCHDOG_USEC doen't exist FAILED")
	}

	// (0, nil) WATCHDOG_PID doen't exist
	err = os.Setenv("WATCHDOG_USEC", "1")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("WATCHDOG_PID")
	if err != nil {
		panic(err)
	}
	delay, err = SdWatchdogEnabled()
	if delay != 0 || err != nil {
		t.Errorf("TEST: WATCHDOG_PID doen't exist FAILED")
	}

	// (0, err) USEC negative
	err = os.Setenv("WATCHDOG_USEC", "-1")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("WATCHDOG_PID", strconv.Itoa(os.Getpid()))
	if err != nil {
		panic(err)
	}
	_, err = SdWatchdogEnabled()
	if err == nil {
		t.Errorf("TEST: USEC negative FAILED")
	}

	// (0, err) Bad USEC
	err = os.Setenv("WATCHDOG_USEC", "string")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("WATCHDOG_PID", strconv.Itoa(os.Getpid()))
	if err != nil {
		panic(err)
	}
	_, err = SdWatchdogEnabled()
	if err == nil {
		t.Errorf("TEST: Bad USEC FAILED")
	}

	// (0, err) Bad PID
	err = os.Setenv("WATCHDOG_USEC", "1")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("WATCHDOG_PID", "string")
	if err != nil {
		panic(err)
	}
	_, err = SdWatchdogEnabled()
	if err == nil {
		t.Errorf("TEST: Bad PID FAILED")
	}
}
