package journal

import (
	"strconv"
	"time"
)

// Value represents a key-value pair that can be logged to the journal.
type Value interface {
	// Name must be composed of uppercase letters, numbers,
	// and underscores, but must not start with an underscore. Within these
	// restrictions, any arbitrary field name may be used.  Some names have special
	// significance: see the journalctl documentation
	// (http://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html)
	// for more details.
	Name() string

	// Value must be a string representation of the value.
	Value() string
}

type value struct {
	name  string
	value string
}

var _ = Value(&value{})

func (s value) Name() string {
	return s.name
}

func (s value) Value() string {
	return s.value
}

// String returns a Value for a string value.
func String(name, Value string) Value {
	return &value{name, Value}
}

// Int returns a Value for an int.
func Int(name string, Value int) Value {
	return &value{name, strconv.Itoa(Value)}
}

// Int64 returns a Value for an int64.
func Int64(name string, Value int64) Value {
	return &value{name, strconv.FormatInt(Value, 10)}
}

// Uint64 returns a Value for a uint64.
func Uint64(name string, Value uint64) Value {
	return &value{name, strconv.FormatUint(Value, 10)}
}

// Float64 returns a Value for a floating-point number.
func Float64(name string, Value float64) Value {
	return &value{name, strconv.FormatFloat(Value, 'g', -1, 64)}
}

// Bool returns a Value for a bool.
func Bool(name string, Value bool) Value {
	return &value{name, strconv.FormatBool(Value)}
}

// Time returns a Value for a time.Time.
func Time(name string, Value time.Time) Value {
	return &value{name, Value.Format(time.RFC3339Nano)}
}

// Duration returns a Value for a time.Duration.
func Duration(name string, Value time.Duration) Value {
	return &value{name, Value.String()}
}

// Error returns a Value for an error.
func Error(name string, Value error) Value {
	return &value{name, Value.Error()}
}
