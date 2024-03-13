package flag

import (
	"fmt"
	"strings"
)

// IntFlag pFlag wrapper
type IntFlag struct {
	Name           string
	Alias          string
	Usage          string
	Value          int
	Required       bool
	Hidden         bool
	PersistentFlag bool
}

func (f *IntFlag) isFlag() bool {
	return true
}

func (f *IntFlag) GetName() string {
	return f.Name
}

func (f *IntFlag) IsRequired() bool {
	return f.Required
}

func (f *IntFlag) IsHidden() bool {
	return f.Hidden
}

func (f *IntFlag) ParseToString() string {
	usage := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
	if f.Required {
		return fmt.Sprintf("--%s %s", f.Name, usage)
	}
	return fmt.Sprintf("[--%s %s]", f.Name, usage)
}

func (f *IntFlag) UniqueKey() string {
	return f.Name // Assuming Name is unique across all flags
}

func (f *IntFlag) IsPersistentFlag() bool {
	return f.PersistentFlag
}
