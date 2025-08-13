// Copyright 2022 CoreOS, Inc.
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
// limitations under the License

package main

import (
	"fmt"
	"os"

	"github.com/coreos/go-systemd/v22/journal"
)

func main() {
	ok, err := journal.StderrIsJournalStream()
	if err != nil {
		panic(err)
	}

	if ok {
		// use journal native protocol
		_ = journal.Send("this is a message logged through the native protocol", journal.PriInfo, nil)
	} else {
		// use stderr
		fmt.Fprintln(os.Stderr, "this is a message logged through stderr")
	}
}
