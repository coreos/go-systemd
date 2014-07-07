package unit

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// DeserializeUnitFile parses a systemd unit file and attempts to map its various sections and values.
// Currently this function is dangerously simple and should be rewritten to match the systemd unit file spec
func DeserializeUnitFile(raw string) (map[string]map[string][]string, error) {
	sections := make(map[string]map[string][]string)
	var section string
	var prev string
	for i, line := range strings.Split(raw, "\n") {

		// Join lines ending in backslash
		if strings.HasSuffix(line, "\\") {
			// Replace trailing slash with space
			prev = prev + line[:len(line)-1] + " "
			continue
		}

		// Concatenate any previous conjoined lines
		if prev != "" {
			line = prev + line
			prev = ""
		} else if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			// Ignore commented-out lines that are not part of a continuation
			continue
		}

		line = strings.TrimSpace(line)

		// Ignore blank lines
		if len(line) == 0 {
			continue
		}

		// Check for section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			sections[section] = make(map[string][]string)
			continue
		}

		// ignore any lines that aren't within a section
		if len(section) == 0 {
			continue
		}

		key, values, err := deserializeUnitLine(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing line %d: %v", i+1, err)
		}
		for _, v := range values {
			sections[section][key] = append(sections[section][key], v)
		}
	}

	return sections, nil
}

func deserializeUnitLine(line string) (key string, values []string, err error) {
	e := strings.Index(line, "=")
	if e == -1 {
		err = errors.New("missing '='")
		return
	}
	key = strings.TrimSpace(line[:e])
	value := strings.TrimSpace(line[e+1:])

	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		for _, v := range parseMultivalueLine(value) {
			values = append(values, v)
		}
	} else {
		values = append(values, value)
	}
	return
}

// parseMultivalueLine parses a line that includes several quoted values separated by whitespaces.
// Example: MachineMetadata="foo=bar" "baz=qux"
func parseMultivalueLine(line string) (values []string) {
	var v bytes.Buffer
	w := false // check whether we're within quotes or not

	for _, e := range []byte(line) {
		// ignore quotes
		if e == '"' {
			w = !w
			continue
		}

		if e == ' ' {
			if !w { // between quoted values, keep the previous value and reset.
				values = append(values, v.String())
				v.Reset()
				continue
			}
		}

		v.WriteByte(e)
	}

	values = append(values, v.String())

	return
}
