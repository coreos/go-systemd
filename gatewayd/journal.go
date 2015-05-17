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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const DefaultHost = "http://localhost:19531"

// A Client is used for accessing the systemd journal through the
// systemd-journal-gatewayd's HTTP API
type Client struct {
	Host string

	http http.Client
}

func (c *Client) host() string {
	if c.Host == "" {
		return DefaultHost
	}
	return c.Host
}

func (c *Client) request(opts *Options) (req *http.Request, err error) {
	req, err = getRequest(c.host(), opts)
	return
}

func getRequest(host string, opts *Options) (req *http.Request, err error) {
	values := getURLParams(opts)
	req, err = http.NewRequest("GET", host+"/entries?"+values.Encode(), nil)
	if err != nil {
		return
	}

	req.Header.Add("Accept", "application/json")
	if opts != nil && opts.Range != nil {
		setRangeHeader(req, *opts.Range)
	}
	return
}

// See http://www.freedesktop.org/software/systemd/man/systemd-journal-gatewayd.service.html#URL%20GET%20parameters
func getURLParams(opts *Options) url.Values {
	values := url.Values{}
	if opts == nil {
		return values
	}
	if opts.Boot {
		values.Set("boot", "true")
	}
	if opts.Follow {
		values.Set("follow", "true")
	}
	if opts.Filter != nil {
		for field, value := range opts.Filter {
			values.Set(field, value)
		}
	}
	return values
}

// See http://www.freedesktop.org/software/systemd/man/systemd-journal-gatewayd.service.html#Range%20header
func setRangeHeader(req *http.Request, rangeOpts RangeOptions) {
	header := fmt.Sprintf("entries=%s:%d:%d", rangeOpts.Cursor, rangeOpts.NumSkip, rangeOpts.NumEntries)
	req.Header.Set("Range", header)
}

type decoder interface {
	Decode(v interface{}) error
}

func DecodeEntry(dec decoder, r io.Reader) (*Entry, error) {
	e := new(Entry)
	if err := dec.Decode(e); err != nil {
		return e, err
	}
	switch t := e.MessageRaw.(type) {
	case []byte:
		e.Message = string(t)
	case string:
		e.Message = t
	// When systemd encodes to JSON, if it finds any non-printable or non-UTF8
	// character it gets serialized to a number array.
	case []interface{}:
		buf := new(bytes.Buffer)
		for _, v := range t {
			item := v.(float64)
			err := binary.Write(buf, binary.BigEndian, item)
			if err != nil {
				return e, err
			}
		}
		e.Message = buf.String()
	}
	return e, nil
}

func newJSONDecoder(r io.Reader) decoder {
	return json.NewDecoder(r)
}
