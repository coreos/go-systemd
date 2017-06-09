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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type RangeOptions struct {
	Cursor     string
	NumSkip    int64
	NumEntries uint64
}

type Options struct {
	Follow bool
	// Discrete bool
	Boot   bool
	Filter map[string]string
	Range  *RangeOptions
}

// Entries will return a list of entries.
func (c *Client) Entries(opts *Options) (entries []*Entry, err error) {
	var (
		req  *http.Request
		resp *http.Response
	)
	req, err = c.request(opts)
	if err != nil {
		return
	}
	resp, err = c.http.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	entries, err = readerToEntries(resp.Body)
	return
}

func (c *Client) Follow(opts *Options) (*Stream, error) {
	if opts == nil {
		opts = &Options{}
	}
	opts.Follow = true
	req, err := c.request(opts)
	if err != nil {
		return nil, err
	}
	s := &Stream{entries: make(chan streamEntry)}
	r, err := s.start(req)
	if err != nil {
		return nil, err
	}
	go s.stream(r)
	return s, nil
}

type streamEntry struct {
	entry *Entry
	err   error
}

type Stream struct {
	http    http.Client
	entries chan streamEntry
}

func (s *Stream) Next() (*Entry, error) {
	se := <-s.entries
	return se.entry, se.err
}

func (s *Stream) start(req *http.Request) (r io.ReadCloser, err error) {
	req.Header.Set("Cache-Control", "no-cache")
	var resp *http.Response
	resp, err = s.http.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = fmt.Errorf(string(body))
		}
		return nil, err
	}
	r = resp.Body
	return
}

func (s *Stream) stream(r io.ReadCloser) {
	defer r.Close()
	decoder := newJSONDecoder(r)

	for {
		entry, err := DecodeEntry(decoder, r)
		s.entries <- streamEntry{entry: entry, err: err}
		if err != nil {
			close(s.entries)
			return
		}
	}
}

func readerToEntries(r io.Reader) (entries []*Entry, err error) {
	decoder := newJSONDecoder(r)

	entries = make([]*Entry, 0)
	for {
		var entry *Entry
		entry, err = DecodeEntry(decoder, r)
		if err == io.EOF {
			err = nil
			return
		} else if err != nil {
			return
		}
		entries = append(entries, entry)
	}
}
