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

package dlopen

import (
	"fmt"
	"sync"
	"testing"
)

func checkFailure(shouldSucceed bool, err error) (rErr error) {
	switch {
	case err != nil && shouldSucceed:
		rErr = fmt.Errorf("expected test to succeed, failed unexpectedly: %v", err)
	case err == nil && !shouldSucceed:
		rErr = fmt.Errorf("expected test to fail, succeeded unexpectedly")
	}

	return
}

func TestDlopen(t *testing.T) {
	tests := []struct {
		libs          []string
		shouldSucceed bool
	}{
		{
			libs: []string{
				"libc.so.6",
				"libc.so",
			},
			shouldSucceed: true,
		},
		{
			libs: []string{
				"libstrange.so",
			},
			shouldSucceed: false,
		},
	}

	for i, tt := range tests {
		expLen := 4
		len, err := strlen(tt.libs, "test")
		if checkFailure(tt.shouldSucceed, err) != nil {
			t.Errorf("case %d: %v", i, err)
		}

		if tt.shouldSucceed && len != expLen {
			t.Errorf("case %d: expected length %d, got %d", i, expLen, len)
		}
	}
}

// Note this is not a reliable reproducer for the problem.
// It depends on the fact the it first generates some dlerror() errors
// by using non existent libraries.
func TestDlopenThreadSafety(t *testing.T) {
	t.Skip("panics in CI; see https://github.com/coreos/go-systemd/issues/462.")
	libs := []string{
		"libstrange1.so",
		"libstrange2.so",
		"libstrange3.so",
		"libstrange4.so",
		"libstrange5.so",
		"libstrange6.so",
		"libstrange7.so",
		"libstrange8.so",
		"libc.so.6",
		"libc.so",
	}

	// 10000 is the default golang thread limit, so adding more will likely fail
	// but this number is enough to reproduce the issue most of the time for me.
	count := 10000
	wg := sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			lib, err := GetHandle(libs)
			if err != nil {
				t.Errorf("GetHandle failed unexpectedly: %v", err)
			}
			_, err = lib.GetSymbolPointer("strlen")
			if err != nil {
				t.Errorf("GetSymbolPointer strlen failed unexpectedly: %v", err)
			}
		}()
	}
	wg.Wait()
}
