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
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

func TestValidVarName(t *testing.T) {
	validTestCases := []string{
		"TEST",
		"TE_ST",
		"TEST_",
		"0TEST0",
	}
	invalidTestCases := []string{
		"test",
		"_TEST",
		"",
	}

	for _, tt := range validTestCases {
		if err := validVarName(tt); err != nil {
			t.Fatalf("\"%s\" should be a valid variable", tt)
		}
	}
	for _, tt := range invalidTestCases {
		if err := validVarName(tt); err == nil {
			t.Fatalf("\"%s\" should be an invalid variable", tt)
		}
	}

}

func TestJournalSend(t *testing.T) {
	// an always-too-big value (hopefully)
	hugeValue := 1234567890

	// a value slightly larger than default limit,
	// see `SO_SNDBUF` in socket(7)
	largeValue := hugeValue
	if wmem, err := ioutil.ReadFile("/proc/sys/net/core/wmem_default"); err == nil {
		wmemStr := strings.TrimSpace(string(wmem))
		if v, err := strconv.Atoi(wmemStr); err == nil {
			largeValue = v + 1
		}
	}
	// See https://github.com/coreos/go-systemd/pull/221#issuecomment-276727718
	_ = largeValue

	// small messages should go over normal data,
	// larger ones over temporary file with fd in ancillary data
	testValues := []struct {
		label string
		len   int
	}{
		{
			"empty message",
			0,
		},
		{
			"small message",
			5,
		},
		/* See https://github.com/coreos/go-systemd/pull/221#issuecomment-276727718
		{
			"large message",
			largeValue,
		},
		{
			"huge message",
			hugeValue,
		},
		*/
	}

	for i, tt := range testValues {
		t.Logf("journal send test #%v - %s (len=%d)", i, tt.label, tt.len)
		largeVars := map[string]string{
			"KEY": string(make([]byte, tt.len)),
		}

		err := Send(fmt.Sprintf("go-systemd test #%v - %s", i, tt.label), PriCrit, largeVars)
		if err != nil {
			t.Fatalf("#%v: failed sending %s: %s", i, tt.label, err)
		}
	}
}
