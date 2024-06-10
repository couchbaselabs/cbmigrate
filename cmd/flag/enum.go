package flag

import (
	"fmt"
	"strings"
)

// EnumFlag pFlag wrapper
type EnumFlag struct {
	Name           string
	Alias          string
	Usage          string
	Values         []string
	DefaultValue   string
	Required       bool
	Hidden         bool
	PersistentFlag bool
}

func (f *EnumFlag) isFlag() bool {
	return true
}

func (f *EnumFlag) GetName() string {
	return f.Name
}

func (f *EnumFlag) IsRequired() bool {
	return f.Required
}

func (f *EnumFlag) IsHidden() bool {
	return f.Hidden
}

func (f *EnumFlag) ParseToString() string {
	usage := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
	if len(f.Values) > 0 {
		usage = strings.Join(f.Values, ",")
	}
	if f.Required {
		return fmt.Sprintf("--%s %s", f.Name, usage)
	}
	return fmt.Sprintf("[--%s %s]", f.Name, usage)
}

func (f *EnumFlag) UniqueKey() string {
	return f.Name // Assuming Name is unique across all flags
}

func (f *EnumFlag) IsPersistentFlag() bool {
	return f.PersistentFlag
}
