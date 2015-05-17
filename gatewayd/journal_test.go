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

package gatewayd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestEntriesSingleEntry(t *testing.T) {

	expected := Entry{
		Message:                 "fedora : TTY=pts/1 ; PWD=/home/fedora ; USER=root ; COMMAND=/bin/systemctl status systemd-journal-gatewayd",
		Cursor:                  "s=c60ae6415a0e4531a2064c8cdb11a0f4;i=587;b=d5f050252af342a2817795b375889f78;m=b11f9d6;t=5120f62678669;x=74d4678ad474a6b6",
		RealtimeTimestamp:       "1427232168314473",
		MonotonicTimestamp:      "185727446",
		BootID:                  "d5f050252af342a2817795b375889f78",
		UID:                     "0",
		MachineID:               "fe39ba83b9244251b1704fc655fbff2f",
		Priority:                "5",
		CapEffective:            "3fffffffff",
		Transport:               "syslog",
		Hostname:                "ip-172-30-0-229",
		SyslogFacility:          "10",
		GID:                     "1000",
		SystemdOwnerUID:         "1000",
		SystemdSlice:            "user-1000.slice",
		AuditLoginUID:           "1000",
		SyslogIdentifier:        "sudo",
		Comm:                    "sudo",
		Exe:                     "/usr/bin/sudo",
		SelinuxContext:          "unconfined_u:unconfined_r:unconfined_t:s0-s0:c0.c1023",
		AuditSession:            "3",
		SystemdCgroup:           "/user.slice/user-1000.slice/session-3.scope",
		SystemdSession:          "3",
		SystemdUnit:             "session-3.scope",
		PID:                     "1125",
		CmdLine:                 "sudo systemctl status systemd-journal-gatewayd",
		SourceRealtimeTimestamp: "1427232168313982",
	}

	bts, err := ioutil.ReadFile("json/journalctl_entry.json")
	if err != nil {
		t.Error(err)
		return
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(bts))
	}))
	defer ts.Close()

	client := &Client{Host: ts.URL}
	entries, err := client.Entries(nil)
	if err != nil {
		t.Error(err)
		return
	}

	numEntries := len(entries)
	if numEntries != 1 {
		t.Fatalf("Expected 1 entry, got %d entries", numEntries)
	}
	entry := entries[0]
	if expected == *entry {
		t.Errorf("Expected and actual entries differ:\n%+v\n%+v", expected, entry)
	}
}

func TestEntriesMultipleEntries(t *testing.T) {
	bts, err := ioutil.ReadFile("json/journalctl_entry_list.json")
	if err != nil {
		t.Error(err)
		return
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(bts))
	}))
	defer ts.Close()
	client := &Client{Host: ts.URL}
	entries, err := client.Entries(nil)
	if err != nil {
		t.Error(err)
		return
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries and found %d", len(entries))
		return
	}

	expected, err := readerToEntries(bytes.NewBuffer(bts))
	if err != nil {
		t.Fatalf("Exected no err, got err=%v", err)
	}

	for i, entry := range entries {
		if !reflect.DeepEqual(entry, expected[i]) {
			t.Errorf("Expected and actual entries differ:\n%+v\n%+v", expected[i], entry)
		}
	}
}

func TestFollowEntries(t *testing.T) {
	bts, err := ioutil.ReadFile("json/journalctl_entry_list.json")
	if err != nil {
		t.Error(err)
		return
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		buf := bytes.NewBuffer(bts)
		for {
			line, err := buf.ReadBytes(byte('\n'))
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			fmt.Fprintf(w, "%s\n", line)
		}
	}))
	defer ts.Close()
	client := &Client{Host: ts.URL}
	stream, err := client.Follow(nil)
	if err != nil {
		t.Error(err)
		return
	}

	entries := make([]*Entry, 0)
	for {
		entry, err := stream.Next()
		if err != nil {
			break
		}
		entries = append(entries, entry)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries and found %d", len(entries))
		return
	}

	expected, err := readerToEntries(bytes.NewBuffer(bts))
	if err != nil {
		t.Fatalf("Exected no err, got err=%v", err)
	}

	for i, entry := range entries {
		if !reflect.DeepEqual(entry, expected[i]) {
			t.Errorf("Expected and actual entries differ:\n%+v\n%+v", expected[i], entry)
		}
	}
}
