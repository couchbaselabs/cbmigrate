package flag

import (
	"fmt"
)

// BoolFlag pFlag wrapper
type BoolFlag struct {
	Name           string
	Alias          string
	Usage          string
	Value          bool
	Required       bool
	Hidden         bool
	PersistentFlag bool
}

func (f *BoolFlag) isFlag() bool {
	return true
}

func (f *BoolFlag) IsPersistentFlag() bool {
	return f.PersistentFlag
}

// GetName get flag name
func (f *BoolFlag) GetName() string {
	return f.Name
}

// IsRequired Flag required or not
func (f *BoolFlag) IsRequired() bool {
	return f.Required
}

// IsHidden Flag hidden or not
func (f *BoolFlag) IsHidden() bool {
	return f.Hidden
}

func (f *BoolFlag) ParseToString() string {
	if f.Required {
		return fmt.Sprintf("--%s", f.Name)
	}
	return fmt.Sprintf("[--%s]", f.Name)
}

func (f *BoolFlag) UniqueKey() string {
	return f.Name // Assuming Name is unique across all flags
}
