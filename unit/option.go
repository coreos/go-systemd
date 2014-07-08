package unit

import (
	"fmt"
)

type UnitOption struct {
	Section string
	Name    string
	Value   string
}

func (uo *UnitOption) String() string {
	return fmt.Sprintf("{Section: %q, Name: %q, Value: %q}", uo.Section, uo.Name, uo.Value)
}
