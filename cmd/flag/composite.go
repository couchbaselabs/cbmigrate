package flag

import (
	"fmt"
	"strings"
)

type CompositeFlagType uint

const (
	RelationshipAND CompositeFlagType = iota
	RelationshipOR
)

// CompositeFlag struct for composite flags
type CompositeFlag struct {
	Flags         []Flag
	Required      bool
	Hidden        bool
	RequiredBrace bool
	Type          CompositeFlagType
}

func (cf *CompositeFlag) isFlag() bool {
	return true
}

func (cf *CompositeFlag) GetName() string {
	return ""
}

func (cf *CompositeFlag) IsRequired() bool {
	return cf.Required
}

func (cf *CompositeFlag) IsHidden() bool {
	return cf.Hidden
}

// ParseToString method for CompositeFlag
func (cf *CompositeFlag) ParseToString() string {
	var parts []string
	for _, flag := range cf.Flags {
		if !flag.IsHidden() {
			parts = append(parts, flag.ParseToString())
		}
	}
	sperator := " "
	if cf.Type == RelationshipOR {
		sperator = " | "
	}
	joined := strings.Join(parts, sperator)

	switch {
	case cf.RequiredBrace:
		return fmt.Sprintf("(%s)", joined)
	case cf.Required:
		return fmt.Sprintf("%s", joined)
	}
	return fmt.Sprintf("[%s]", joined)
}

func (cf *CompositeFlag) UniqueKey() string {
	key := "composite:"
	for _, flag := range cf.Flags {
		key += flag.UniqueKey() + ";"
	}
	return key
}
