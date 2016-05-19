// Copyright 2015 RedHat, Inc.
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

package sdjournal

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"golang.org/x/net/context"
)

var (
	ErrExpired = errors.New("Timeout expired")
)

// JournalReaderConfig represents options to drive the behavior of a JournalReader.
type JournalReaderConfig struct {
	// The Since and NumFromTail options are mutually exclusive and determine
	// where the reading begins within the journal.
	Since       time.Duration // start relative to a Duration from now
	NumFromTail uint64        // start relative to the tail

	// Show only journal entries whose fields match the supplied values. If
	// the array is empty, entries will not be filtered.
	Matches []Match

	// If not empty, the journal instance will point to a journal residing
	// in this directory. The supplied path may be relative or absolute.
	Path string
}

// JournalReader is an io.ReadCloser which provides a simple interface for iterating through the
// systemd journal.
type JournalReader struct {
	journal *Journal
}

// FollowFilter is a function which you pass into Follow to determine if you to
// filter key value pair from a journal entry.
type FollowFilter func(key, value string) bool

// AllKeys is a FollowFilter that allows all keys of a given journal entry through the filter
func AllKeys(key, value string) bool {
	return true
}

func OnlyMessages(key, value string) bool {
	if key == "MESSAGE" {
		return true
	}
	return false
}

// NewJournalReader creates a new JournalReader with configuration options that are similar to the
// systemd journalctl tool's iteration and filtering features.
func NewJournalReader(config JournalReaderConfig) (*JournalReader, error) {
	r := &JournalReader{}

	// Open the journal
	var err error
	if config.Path != "" {
		r.journal, err = NewJournalFromDir(config.Path)
	} else {
		r.journal, err = NewJournal()
	}
	if err != nil {
		return nil, err
	}

	// Add any supplied matches
	for _, m := range config.Matches {
		r.journal.AddMatch(m.String())
	}

	// Set the start position based on options
	if config.Since != 0 {
		// Start based on a relative time
		start := time.Now().Add(config.Since)
		if err := r.journal.SeekRealtimeUsec(uint64(start.UnixNano() / 1000)); err != nil {
			return nil, err
		}
	} else if config.NumFromTail != 0 {
		// Start based on a number of lines before the tail
		if err := r.journal.SeekTail(); err != nil {
			return nil, err
		}

		// Move the read pointer into position near the tail. Go one further than
		// the option so that the initial cursor advancement positions us at the
		// correct starting point.
		if _, err := r.journal.PreviousSkip(config.NumFromTail + 1); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *JournalReader) Read(b []byte) (int, error) {
	var err error
	var c int

	// Advance the journal cursor
	c, err = r.journal.Next()

	// An unexpected error
	if err != nil {
		return 0, err
	}

	// EOF detection
	if c == 0 {
		return 0, io.EOF
	}

	// Build a message
	var msg string
	msg, err = r.buildMessage()

	if err != nil {
		return 0, err
	}

	// Copy and return the message
	copy(b, []byte(msg))

	return len(msg), nil
}

func (r *JournalReader) Close() error {
	return r.journal.Close()
}

// Follow asynchronously follows the JournalReader it takes in a context to stop the following the Journal, it returns a
// buffered channel of errors and will stop following the journal on the first given error
func (r *JournalReader) Follow(ctx context.Context, msgs chan<- map[string]interface{}, filter FollowFilter) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		for {
			kvMap := make(map[string]interface{})
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				var err error
				var c int

				// Advance the journal cursor
				c, err = r.journal.Next()

				// An unexpected error
				if err != nil {
					errChan <- err
					return
				}

				// We have a new journal entry go over the fields
				// get the data for what we care about and return
				r.journal.RestartData()
				if c > 0 {
				fields:
					for {
						s, err := r.journal.EnumerateData()
						if err != nil || len(s) == 0 {
							break fields
						}
						s = s[:len(s)]
						arr := strings.SplitN(s, "=", 2)
						// if we want the pair,
						// add it to the map
						if filter(arr[0], arr[1]) {
							kvMap[arr[0]] = arr[1]
						}
					}
					msgs <- kvMap
				}
			}

			// we're at the tail, so wait for new events or time out.
			// holds journal events to process. tightly bounded for now unless there's a
			// reason to unblock the journal watch routine more quickly
			events := make(chan int, 1)
			pollDone := make(chan bool, 1)
			go func() {
				for {
					select {
					case <-pollDone:
						return
					default:
						events <- r.journal.Wait(time.Duration(1) * time.Second)
					}
				}
			}()

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				pollDone <- true
				return
			case e := <-events:
				pollDone <- true
				switch e {
				case SD_JOURNAL_NOP, SD_JOURNAL_APPEND, SD_JOURNAL_INVALIDATE:
					// TODO: need to account for any of these?
				default:
					log.Printf("Received unknown event: %d\n", e)
				}
			}
		}
	}()
	return errChan
}

// buildMessage returns a string representing the current journal entry in a simple format which
// includes the entry timestamp and MESSAGE field.
func (r *JournalReader) buildMessage() (string, error) {
	var msg string
	var usec uint64
	var err error

	if msg, err = r.journal.GetData("MESSAGE"); err != nil {
		return "", err
	}

	if usec, err = r.journal.GetRealtimeUsec(); err != nil {
		return "", err
	}

	timestamp := time.Unix(0, int64(usec)*int64(time.Microsecond))

	return fmt.Sprintf("%s %s\n", timestamp, msg), nil
}
