package flag

import (
	"fmt"
	"strings"
)

// StringSliceFlag pFlag wrapper
type StringSliceFlag struct {
	Name     string
	Alias    string
	Usage    string
	Value    []string
	Required bool
	Hidden   bool
}

func (f *StringSliceFlag) isFlag() bool {
	return true
}

func (f *StringSliceFlag) GetName() string {
	return f.Name
}

func (f *StringSliceFlag) IsRequired() bool {
	return f.Required
}

func (f *StringSliceFlag) IsHidden() bool {
	return f.Hidden
}

func (f *StringSliceFlag) ParseToString() string {
	usage := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
	if f.Required {
		return fmt.Sprintf("--%s %s", f.Name, usage)
	}
	return fmt.Sprintf("[--%s %s]", f.Name, usage)
}

func (f *StringSliceFlag) UniqueKey() string {
	return f.Name // Assuming Name is unique across all flags
}
