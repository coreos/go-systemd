package unit

import (
	"reflect"
	"testing"
)

func TestDeserializeFile(t *testing.T) {
	contents := `
This=Ignored
[Unit]
;ignore this guy
Description = Foo

[Service]
ExecStart=echo "ping";
ExecStop=echo "pong"
# ignore me, too
ExecStop=echo post

[Fleet]
X-ConditionMachineMetadata=foo=bar
X-ConditionMachineMetadata=baz=qux
`

	expected := map[string]map[string][]string{
		"Unit": {
			"Description": {"Foo"},
		},
		"Service": {
			"ExecStart": {`echo "ping";`},
			"ExecStop":  {`echo "pong"`, "echo post"},
		},
		"Fleet": {
			"X-ConditionMachineMetadata": {"foo=bar", "baz=qux"},
		},
	}

	unitFile, err := DeserializeUnitFile(contents)
	if err != nil {
		t.Fatalf("Unexpected error parsing unit %q: %v", contents, err)
	}

	if !reflect.DeepEqual(expected, unitFile) {
		t.Fatalf("Map func did not produce expected output.\nActual=%v\nExpected=%v", unitFile, expected)
	}
}

func TestDeserializedUnitFileGarbage(t *testing.T) {
	contents := `
>>>>>>>>>>>>>
[Service]
ExecStart=jim
# As long as a line has an equals sign, systemd is happy, so we should pass it through
<<<<<<<<<<<=bar
`
	expected := map[string]map[string][]string{
		"Service": {
			"ExecStart":   {"jim"},
			"<<<<<<<<<<<": {"bar"},
		},
	}
	unitFile, err := DeserializeUnitFile(contents)
	if err != nil {
		t.Fatalf("Unexpected error parsing unit %q: %v", contents, err)
	}

	if !reflect.DeepEqual(expected, unitFile) {
		t.Fatalf("Map func did not produce expected output.\nActual=%v\nExpected=%v", unitFile, expected)
	}
}

func TestDeserializeUnitFileEscapedMultilines(t *testing.T) {
	contents := `
[Service]
ExecStart=echo \
  "pi\
  ng"
ExecStop=\
echo "po\
ng"
# comments within continuation should not be ignored
ExecStopPre=echo\
#pang
ExecStopPost=echo\
#peng\
pung
`
	expected := map[string]map[string][]string{
		"Service": {
			"ExecStart":    {`echo    "pi   ng"`},
			"ExecStop":     {`echo "po ng"`},
			"ExecStopPre":  {`echo #pang`},
			"ExecStopPost": {`echo #peng pung`},
		},
	}
	unitFile, err := DeserializeUnitFile(contents)
	if err != nil {
		t.Fatalf("Unexpected error parsing unit %q: %v", contents, err)
	}

	if !reflect.DeepEqual(expected, unitFile) {
		t.Fatalf("Map func did not produce expected output.\nActual=%v\nExpected=%v", unitFile, expected)
	}
}

func TestDeserializeLine(t *testing.T) {
	deserializeLineExamples := map[string][]string{
		`key=foo=bar`:             {`foo=bar`},
		`key="foo=bar"`:           {`foo=bar`},
		`key="foo=bar" "baz=qux"`: {`foo=bar`, `baz=qux`},
		`key="foo=bar baz"`:       {`foo=bar baz`},
		`key="foo=bar" baz`:       {`"foo=bar" baz`},
		`key=baz "foo=bar"`:       {`baz "foo=bar"`},
		`key="foo=bar baz=qux"`:   {`foo=bar baz=qux`},
	}

	for q, w := range deserializeLineExamples {
		k, g, err := deserializeUnitLine(q)
		if err != nil {
			t.Fatalf("Unexpected error testing %q: %v", q, err)
		}
		if k != "key" {
			t.Fatalf("Unexpected key, got %q, want %q", k, "key")
		}
		if !reflect.DeepEqual(g, w) {
			t.Errorf("Unexpected line parse for %q:\ngot %q\nwant %q", q, g, w)
		}
	}

	// Any non-empty line without an '=' is bad
	badLines := []string{
		`<<<<<<<<<<<<<<<<<<<<<<<<`,
		`asdjfkl;`,
		`>>>>>>>>>>>>>>>>>>>>>>>>`,
		`!@#$%^&&*`,
	}
	for _, l := range badLines {
		_, _, err := deserializeUnitLine(l)
		if err == nil {
			t.Fatalf("Did not get expected error deserializing %q", l)
		}
	}
}

func TestBadUnitsFail(t *testing.T) {
	bad := []string{
		`
[Unit]

[Service]
<<<<<<<<<<<<<<<<
`,
		`
[Unit]
nonsense upon stilts
`,
	}
	for _, tt := range bad {
		if _, err := DeserializeUnitFile(tt); err == nil {
			t.Fatalf("Did not get expected error creating Unit from %q", tt)
		}
	}
}
