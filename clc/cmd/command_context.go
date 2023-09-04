package cmd

import (
	"fmt"
	"math"

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type ArgType int

const (
	ArgTypeString ArgType = iota
	ArgTypeStringSlice
)

type ArgSpec struct {
	Key   string
	Title string
	Type  ArgType
	Min   int
	Max   int
}

type CommandContext struct {
	Cmd          *cobra.Command
	CP           config.Provider
	stringValues map[string]*string
	boolValues   map[string]*bool
	intValues    map[string]*int64
	mode         Mode
	isTopLevel   bool
	group        *cobra.Group
	argSpecs     []ArgSpec
}

func NewCommandContext(cmd *cobra.Command, cfgProvider config.Provider, mode Mode) *CommandContext {
	return &CommandContext{
		Cmd:          cmd,
		CP:           cfgProvider,
		stringValues: map[string]*string{},
		boolValues:   map[string]*bool{},
		intValues:    map[string]*int64{},
		mode:         mode,
	}
}

func (cc *CommandContext) AddStringFlag(long, short, value string, required bool, help string) {
	var s string
	cc.Cmd.PersistentFlags().StringVarP(&s, long, short, value, help)
	if required {
		check.Must(cc.Cmd.MarkPersistentFlagRequired(long))
	}
	cc.stringValues[long] = &s
}

func (cc *CommandContext) AddIntFlag(long, short string, value int64, required bool, help string) {
	var i int64
	cc.Cmd.PersistentFlags().Int64VarP(&value, long, short, value, help)
	if required {
		check.Must(cc.Cmd.MarkPersistentFlagRequired(long))
	}
	cc.intValues[long] = &i
}

func (cc *CommandContext) AddBoolFlag(long, short string, value bool, required bool, help string) {
	cc.Cmd.AddGroup()
	var b bool
	cc.Cmd.PersistentFlags().BoolVarP(&b, long, short, value, help)
	if required {
		check.Must(cc.Cmd.MarkPersistentFlagRequired(long))
	}
	cc.boolValues[long] = &b
}

func (cc *CommandContext) AddStringArg(key, title, help string) {
	s := ArgSpec{
		Key:   key,
		Title: title,
		Type:  ArgTypeString,
		Min:   1,
		Max:   1,
	}
	cc.argSpecs = append(cc.argSpecs, s)
}

// SetPositionalArgCount sets the number minimum and maximum positional arguments.
// if min and max are the same, the pos args are set as the exact num of args.
// otherwise, if max == math.MaxInt, num of pos args are set as the minumum of min args.
// otherwise, if min == 0, num of pos args are set as the maximum of max args.
// otherwise num of pos args is the range of min, max args.
func (cc *CommandContext) SetPositionalArgCount(min, max int) {
	if min == max {
		cc.Cmd.Args = cobra.ExactArgs(min)
		return
	}
	if max == math.MaxInt {
		cc.Cmd.Args = cobra.MinimumNArgs(min)
		return
	}
	if min == 0 {
		cc.Cmd.Args = cobra.MaximumNArgs(max)
		return
	}
	cc.Cmd.Args = cobra.RangeArgs(min, max)
}

func (cc *CommandContext) Hide() {
	cc.Cmd.Hidden = true
}

func (cc *CommandContext) Interactive() bool {
	return cc.mode == ModeInteractive
}

func (cc *CommandContext) SetCommandHelp(long, short string) {
	if long != "" {
		cc.Cmd.Long = long
	}
	if short != "" {
		cc.Cmd.Short = short
	}
}

func (cc *CommandContext) SetCommandUsage(usage string) {
	cc.Cmd.Use = usage
}

func (cc *CommandContext) SetCommandGroup(id string) {
	cc.Cmd.GroupID = id
}

func (cc *CommandContext) AddCommandGroup(id, title string) {
	cc.group = &cobra.Group{
		ID:    id,
		Title: title,
	}
}

func (cc *CommandContext) Group() *cobra.Group {
	return cc.group
}

func (cc *CommandContext) AddStringConfig(name, value, flag string, help string) {
	cc.CP.Set(name, value)
	if flag != "" && !cc.Interactive() {
		f := cc.Cmd.Flag(flag)
		if f != nil {
			cc.CP.BindFlag(name, f)
		}
	}
}

func (cc *CommandContext) SetTopLevel(b bool) {
	cc.isTopLevel = b
}

func (cc *CommandContext) TopLevel() bool {
	return cc.isTopLevel
}

func (cc *CommandContext) ArgsFunc() func(*cobra.Command, []string) error {
	if len(cc.argSpecs) == 0 {
		return cobra.NoArgs
	}
	// validate specs
	for i, s := range cc.argSpecs {
		// min and max should be 1 if this is not the last argspec
		if i == len(cc.argSpecs)-1 {
			if s.Min != 1 || s.Max != 1 {
				panic("only the last argument may take a range of values")
			}
		}
	}
	fn := func(_ *cobra.Command, args []string) error {
		var minIdx int
		c := len(args)
		for _, s := range cc.argSpecs {
			if c < minIdx+s.Min {
				return fmt.Errorf("%s is required", s.Title)
			}
			minIdx += s.Min
		}
		return nil
	}
	return fn
}
