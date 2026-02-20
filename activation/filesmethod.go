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

package activation

// Method decides what happens to the file descriptors that are passed in by systemd.
type Method int

const (
	// ConsumeFiles is the default, and removes the original file descriptors passed in by
	// systemd. This means that new file descriptors created by the program may use the
	// file descriptors indices.
	ConsumeFiles Method = iota

	// ReserveFiles stores placeholder file descriptors, which point to /dev/null. This
	// stops new file descriptors consuming the indices.
	ReserveFiles

	// CloneFiles duplicates and leaves the original file descriptors in tack. This
	// stops new file descriptors consuming the indices like [ReserveFiles] but
	// consider the possible secuirty risk of leaving the sockets exposed.
	ConserveFiles
)
