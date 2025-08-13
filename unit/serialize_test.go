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

package unit

import (
	"io"
	"testing"
)

func TestSerialize(t *testing.T) {
	tests := []struct {
		input  []*UnitOption
		output string
	}{
		// no options results in empty file
		{
			[]*UnitOption{},
			``,
		},

		// options with same section share the header
		{
			[]*UnitOption{
				{"Unit", "Description", "Foo"},
				{"Unit", "BindsTo", "bar.service"},
			},
			`[Unit]
Description=Foo
BindsTo=bar.service
`,
		},

		// options with same name are not combined
		{
			[]*UnitOption{
				{"Unit", "Description", "Foo"},
				{"Unit", "Description", "Bar"},
			},
			`[Unit]
Description=Foo
Description=Bar
`,
		},

		// multiple options printed under different section headers
		{
			[]*UnitOption{
				{"Unit", "Description", "Foo"},
				{"Service", "ExecStart", "/usr/bin/sleep infinity"},
			},
			`[Unit]
Description=Foo

[Service]
ExecStart=/usr/bin/sleep infinity
`,
		},

		// options are grouped into sections
		{
			[]*UnitOption{
				{"Unit", "Description", "Foo"},
				{"Service", "ExecStart", "/usr/bin/sleep infinity"},
				{"Unit", "BindsTo", "bar.service"},
			},
			`[Unit]
Description=Foo
BindsTo=bar.service

[Service]
ExecStart=/usr/bin/sleep infinity
`,
		},

		// options are ordered within groups, and sections are ordered in the order in which they were first seen
		{
			[]*UnitOption{
				{"Unit", "Description", "Foo"},
				{"Service", "ExecStart", "/usr/bin/sleep infinity"},
				{"Unit", "BindsTo", "bar.service"},
				{"X-Foo", "Bar", "baz"},
				{"Service", "ExecStop", "/usr/bin/sleep 1"},
				{"Unit", "Documentation", "https://foo.com"},
			},
			`[Unit]
Description=Foo
BindsTo=bar.service
Documentation=https://foo.com

[Service]
ExecStart=/usr/bin/sleep infinity
ExecStop=/usr/bin/sleep 1

[X-Foo]
Bar=baz
`,
		},

		// utf8 characters are not a problem
		{
			[]*UnitOption{
				{"©", "µ☃", "ÇôrèÕ$"},
			},
			`[©]
µ☃=ÇôrèÕ$
`,
		},

		// no verification is done on section names
		{
			[]*UnitOption{
				{"Un\nit", "Description", "Foo"},
			},
			`[Un
it]
Description=Foo
`,
		},

		// no verification is done on option names
		{
			[]*UnitOption{
				{"Unit", "Desc\nription", "Foo"},
			},
			`[Unit]
Desc
ription=Foo
`,
		},

		// no verification is done on option values
		{
			[]*UnitOption{
				{"Unit", "Description", "Fo\no"},
			},
			`[Unit]
Description=Fo
o
`,
		},
	}

	for i, tt := range tests {
		outReader := Serialize(tt.input)
		outBytes, err := io.ReadAll(outReader)
		if err != nil {
			t.Errorf("case %d: encountered error while reading output: %v", i, err)
			continue
		}

		output := string(outBytes)
		if tt.output != output {
			t.Errorf("case %d: incorrect output", i)
			t.Logf("Expected:\n%s", tt.output)
			t.Logf("Actual:\n%s", output)
		}
	}
}

// TestSerializeSections - test just UnitSecton specific serialization.
func TestSerializeSection(t *testing.T) {
	tests := []struct {
		input  []*UnitSection
		output string
	}{
		// no options results in empty file
		{
			[]*UnitSection{},
			``,
		},

		// options with same section share the header
		{
			[]*UnitSection{
				{
					Section: "Unit",
					Entries: []*UnitEntry{
						{"Description", "Foo"},
						{"BindsTo", "bar.service"},
					},
				},
			},
			`[Unit]
Description=Foo
BindsTo=bar.service
`,
		},

		// options with same name are not combined
		{
			[]*UnitSection{
				{
					Section: "Unit",
					Entries: []*UnitEntry{
						{"Description", "Foo"},
						{"Description", "Bar"},
					},
				},
			},
			`[Unit]
Description=Foo
Description=Bar
`,
		},

		// multiple options printed under different section headers
		{
			[]*UnitSection{
				{
					Section: "Unit",
					Entries: []*UnitEntry{
						{"Description", "Foo"},
					},
				},
				{
					Section: "Service",
					Entries: []*UnitEntry{
						{"ExecStart", "/usr/bin/sleep infinity"},
					},
				},
			},
			`[Unit]
Description=Foo

[Service]
ExecStart=/usr/bin/sleep infinity
`,
		},

		// utf8 characters are not a problem
		{
			[]*UnitSection{
				{
					Section: "©",
					Entries: []*UnitEntry{
						{"µ☃", "ÇôrèÕ$"},
					},
				},
			},
			`[©]
µ☃=ÇôrèÕ$
`,
		},

		// no verification is done on section names
		{
			[]*UnitSection{
				{
					Section: "Un\nit",
					Entries: []*UnitEntry{
						{"Description", "Foo"},
					},
				},
			},
			`[Un
it]
Description=Foo
`,
		},

		// no verification is done on option names
		{
			[]*UnitSection{
				{
					Section: "Unit",
					Entries: []*UnitEntry{
						{"Desc\nription", "Foo"},
					},
				},
			},
			`[Unit]
Desc
ription=Foo
`,
		},

		// no verification is done on option values
		{
			[]*UnitSection{
				{
					Section: "Unit",
					Entries: []*UnitEntry{
						{"Description", "Fo\no"},
					},
				},
			},
			`[Unit]
Description=Fo
o
`,
		},

		// Duplicate sections are preserved.

		{
			[]*UnitSection{
				{
					Section: "Route",
					Entries: []*UnitEntry{
						{"Gateway", "10.0.10.1"},
						{"Destination", "10.0.1.1/24"},
					},
				},
				{
					Section: "Route",
					Entries: []*UnitEntry{
						{"Gateway", "10.0.10.2"},
						{"Destination", "10.0.2.1/24"},
					},
				},
			},
			`[Route]
Gateway=10.0.10.1
Destination=10.0.1.1/24

[Route]
Gateway=10.0.10.2
Destination=10.0.2.1/24
`,
		},
	}

	for i, tt := range tests {
		outReader := SerializeSections(tt.input)
		outBytes, err := io.ReadAll(outReader)
		if err != nil {
			t.Errorf("case %d: encountered error while reading output: %v", i, err)
			continue
		}

		output := string(outBytes)
		if tt.output != output {
			t.Errorf("case %d: incorrect output", i)
			t.Logf("Expected:\n%s", tt.output)
			t.Logf("Actual:\n%s", output)
		}
	}
}
