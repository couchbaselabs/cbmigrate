package flag

import (
	"fmt"
	"strings"
)

// Int64Flag pFlag wrapper
type Int64Flag struct {
	Name     string
	Alias    string
	Usage    string
	Value    int64
	Required bool
	Hidden   bool
}

func (f *Int64Flag) isFlag() bool {
	return true
}

func (f *Int64Flag) GetName() string {
	return f.Name
}

func (f *Int64Flag) IsRequired() bool {
	return f.Required
}

func (f *Int64Flag) IsHidden() bool {
	return f.Hidden
}

func (f *Int64Flag) ParseToString() string {
	usage := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
	if f.Required {
		return fmt.Sprintf("--%s %s", f.Name, usage)
	}
	return fmt.Sprintf("[--%s %s]", f.Name, usage)
}

func (f *Int64Flag) UniqueKey() string {
	return f.Name // Assuming Name is unique across all flags
}
