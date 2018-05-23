// Copyright 2018 CoreOS, Inc.
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

package journal

import (
	"testing"
)

func TestValidaVarName(t *testing.T) {
	testCases := []struct {
		testcase string
		valid    bool
	}{
		{
			"TEST",
			true,
		},
		{
			"test",
			false,
		},
	}

	for _, tt := range testCases {
		valid := validVarName(tt.testcase)
		if valid != tt.valid {
			t.Fatalf("expected %t, got %t", tt.valid, valid)
		}
	}
}
