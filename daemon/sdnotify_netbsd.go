// Copyright 2026
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

//go:build netbsd

package daemon

// SdNotifyMonotonicUsec returns the empty string on NetBSD.
//
// golang.org/x/sys/unix does not export CLOCK_MONOTONIC there, and the
// version required by this module does not export ClockGettime either.
// Callers treat an empty string as "monotonic clock not available".
func SdNotifyMonotonicUsec() string {
	return ""
}
