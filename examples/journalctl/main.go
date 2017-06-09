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

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/gatewayd"
)

var (
	follow     = flag.Bool("follow", false, "Follow logs")
	boot       = flag.Bool("boot", false, "Show only logs for current boot")
	num        = flag.Uint64("n", 10, "Number of logs to grab. 0 for all")
	identifier = flag.String("t", "", "Filter based on this syslog identifier")
)

func main() {
	flag.Parse()
	args := flag.Args()
	var host string
	if len(args) > 0 {
		host = args[0]
	}

	client := &gatewayd.Client{Host: host}
	opts := &gatewayd.Options{
		Boot: *boot,
	}
	if *identifier != "" {
		opts.Filter = map[string]string{
			"SYSLOG_IDENTIFIER": *identifier,
		}
	}
	if *follow {
		stream, err := client.Follow(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error ocurred: %v\n", err)
			os.Exit(1)
		}
		for {
			entry, err := stream.Next()
			if err == io.EOF {
				return
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error occurred getting entry from stream: %v\n", err)
				return
			}
			printEntry(os.Stdout, entry)
		}
	} else {
		if *num != 0 {
			opts.Range = &gatewayd.RangeOptions{
				NumEntries: *num,
			}
		}
		entries, err := client.Entries(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error ocurred: %v\n", err)
			os.Exit(1)
		}
		for _, entry := range entries {
			printEntry(os.Stdout, entry)
		}
	}
}

func printEntry(w io.Writer, entry *gatewayd.Entry) {
	ts, err := strconv.ParseInt(entry.RealtimeTimestamp, 10, 64)
	if err != nil {
		fmt.Fprintf(w, "err: %s\n", err)
		return
	}
	origin := entry.Comm
	if origin == "" {
		origin = entry.Transport
	} else {
		origin += "[" + entry.PID + "]"
	}
	i := int64(ts / (1000 * 1000))
	t := time.Unix(i, 0).UTC().Format("Jan 2 15:04:05")
	fmt.Fprintf(w, "%s %s %s: %s\n", t, entry.Hostname, origin, entry.Message)
}
