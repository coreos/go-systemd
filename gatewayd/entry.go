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

type Entry struct {
	Message                 string
	MessageRaw              interface{} `json:"MESSAGE"`
	Priority                string      `json:"PRIORITY"`
	SyslogFacility          string      `json:"SYSLOG_FACILITY"`
	SyslogIdentifier        string      `json:"SYSLOG_IDENTIFIER"`
	AuditLoginUID           string      `json:"_AUDIT_LOGINUID"`
	AuditSession            string      `json:"_AUDIT_SESSION"`
	BootID                  string      `json:"_BOOT_ID"`
	CapEffective            string      `json:"_CAP_EFFECTIVE"`
	CmdLine                 string      `json:"_CMDLINE"`
	Comm                    string      `json:"_COMM"`
	Exe                     string      `json:"_EXE"`
	GID                     string      `json:"_GID"`
	Hostname                string      `json:"_HOSTNAME"`
	MachineID               string      `json:"_MACHINE_ID"`
	PID                     string      `json:"_PID"`
	SelinuxContext          string      `json:"_SELINUX_CONTEXT"`
	SourceRealtimeTimestamp string      `json:"_SOURCE_REALTIME_TIMESTAMP"`
	SystemdCgroup           string      `json:"_SYSTEMD_CGROUP"`
	SystemdOwnerUID         string      `json:"_SYSTEMD_OWNER_UID"`
	SystemdSession          string      `json:"_SYSTEMD_SESSION"`
	SystemdSlice            string      `json:"_SYSTEMD_SLICE"`
	SystemdUnit             string      `json:"_SYSTEMD_UNIT"`
	Transport               string      `json:"_TRANSPORT"`
	UID                     string      `json:"_UID"`
	Cursor                  string      `json:"__CURSOR"`
	MonotonicTimestamp      string      `json:"__MONOTONIC_TIMESTAMP"`
	RealtimeTimestamp       string      `json:"__REALTIME_TIMESTAMP"`
}
