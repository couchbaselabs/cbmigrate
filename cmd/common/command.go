package common

import (
	"slices"
	"strings"

	"github.com/couchbaselabs/cbmigrate/cmd/flag"
	"github.com/spf13/cobra"
)

type Command string

const (
	Mongo     = "mongo"
	DynamoDB  = "dynamodb"
	CBMigrate = "cbmigrate"
)

var BetaCommands = []Command{
	Mongo,
}

// Example command example
type Example struct {
	Value string
	Usage string
}

// NewCommand command constructor
func NewCommand(name Command, alias []string, examples []Example, short string, long string, flags []flag.Flag) *cobra.Command {

	var example strings.Builder
	exsLen := len(examples)
	for i, ex := range examples {
		example.WriteString("  ")
		example.WriteString(ex.Value)
		if ex.Usage != "" {
			example.WriteString("\n")
			example.WriteString("  ")
			example.WriteString(ex.Usage)
		}
		if i != exsLen-1 {
			example.WriteString("\n")
		}
	}

	var flagUsages []string
	for _, fi := range flags {
		if !fi.IsHidden() {
			flagUsages = append(flagUsages, fi.ParseToString())
		}
	}
	flagUsages = append(flagUsages, "[--help HELP]")
	flagUsage := strings.Join(flagUsages, " ")

	if slices.Index(BetaCommands, name) > -1 {
		if short != "" {
			short = "[Beta] " + short
		}
		if long != "" {
			long = "[Beta] " + long
		}
	}

	cmd := &cobra.Command{
		Use:                   string(name) + " " + flagUsage,
		Aliases:               alias,
		Example:               example.String(),
		Short:                 short,
		Long:                  long,
		DisableFlagsInUseLine: true,
	}
	flags = FlattenFlags(flags)
	for _, fi := range flags {
		flagSet := cmd.Flags()
		switch f := fi.(type) {
		case *flag.IntFlag:
			if f.IsPersistentFlag() {
				flagSet = cmd.PersistentFlags()
			}
			flagSet.IntP(f.Name, f.Alias, f.Value, f.Usage)
		case *flag.Int64Flag:
			if f.IsPersistentFlag() {
				flagSet = cmd.PersistentFlags()
			}
			flagSet.Int64P(f.Name, f.Alias, f.Value, f.Usage)
		case *flag.StringFlag:
			if f.IsPersistentFlag() {
				flagSet = cmd.PersistentFlags()
			}
			flagSet.StringP(f.Name, f.Alias, f.Value, f.Usage)
		case *flag.StringSliceFlag:
			if f.IsPersistentFlag() {
				flagSet = cmd.PersistentFlags()
			}
			flagSet.StringSliceP(f.Name, f.Alias, f.Value, f.Usage)
		case *flag.BoolFlag:
			if f.IsPersistentFlag() {
				flagSet = cmd.PersistentFlags()
			}
			flagSet.BoolP(f.Name, f.Alias, f.Value, f.Usage)
		default:
			panic("flag type not supported for the command " + name)
		}
		if fi.IsHidden() {
			_ = cmd.Flags().MarkHidden(fi.GetName())
		}
	}
	return cmd
}

// FlattenFlags takes a slice of Flag interfaces and returns a deduplicated slice
func FlattenFlags(flags []flag.Flag) []flag.Flag {
	seen := make(map[string]flag.Flag)
	var result []flag.Flag

	for _, f := range flags {
		if _, exists := seen[f.UniqueKey()]; !exists {
			seen[f.UniqueKey()] = f
			// If it's a CompositeFlag, we need to handle its children
			if comp, ok := f.(*flag.CompositeFlag); ok {
				for _, childFlag := range FlattenFlags(comp.Flags) {
					if _, childExists := seen[childFlag.UniqueKey()]; !childExists {
						seen[childFlag.UniqueKey()] = childFlag
						result = append(result, childFlag)
					}
				}
			} else {
				result = append(result, f)
			}
		}
	}
	return result
}
