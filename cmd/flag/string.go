package flag

import (
	"fmt"
	"strings"
)

// StringFlag pFlag wrapper
type StringFlag struct {
	Name     string
	Alias    string
	Usage    string
	Value    string
	Required bool
	Hidden   bool
}

func (f *StringFlag) isFlag() bool {
	return true
}

func (f *StringFlag) GetName() string {
	return f.Name
}

func (f *StringFlag) IsRequired() bool {
	return f.Required
}

func (f *StringFlag) IsHidden() bool {
	return f.Hidden
}

func (f *StringFlag) ParseToString() string {
	usage := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
	if f.Required {
		return fmt.Sprintf("--%s %s", f.Name, usage)
	}
	return fmt.Sprintf("[--%s %s]", f.Name, usage)
}

func (f *StringFlag) UniqueKey() string {
	return f.Name // Assuming Name is unique across all flags
}
