package cmd

import (
	"fmt"
	"math"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type ArgType int

const (
	ArgTypeNone ArgType = iota
	ArgTypeString
	ArgTypeStringSlice
	ArgTypeInt64
	ArgTypeInt64Slice
	ArgTypeKeyValueSlice
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
	usage        string
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

func (cc *CommandContext) AddStringArg(key, title string) {
	s := ArgSpec{
		Key:   key,
		Title: title,
		Type:  ArgTypeString,
		Min:   1,
		Max:   1,
	}
	cc.argSpecs = append(cc.argSpecs, s)
}

func (cc *CommandContext) AddStringSliceArg(key, title string, min, max int) {
	if max < min {
		panic("CommandContext.AddStringSliceArg: max cannot be less than min")
	}
	s := ArgSpec{
		Key:   key,
		Title: title,
		Type:  ArgTypeStringSlice,
		Min:   min,
		Max:   max,
	}
	cc.argSpecs = append(cc.argSpecs, s)
}

func (cc *CommandContext) AddKeyValueSliceArg(key, title string, min, max int) {
	if max < min {
		panic("CommandContext.AddKeyValueSliceArg: max cannot be less than min")
	}
	s := ArgSpec{
		Key:   key,
		Title: title,
		Type:  ArgTypeKeyValueSlice,
		Min:   min,
		Max:   max,
	}
	cc.argSpecs = append(cc.argSpecs, s)
}

func (cc *CommandContext) AddInt64Arg(key, title string) {
	s := ArgSpec{
		Key:   key,
		Title: title,
		Type:  ArgTypeInt64,
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
// Deprecated
func (cc *CommandContext) SetPositionalArgCount(min, max int) {
	// nop
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
	cc.usage = usage
}

func (cc *CommandContext) GetCommandUsage() string {
	return makeCommandUsageString(cc.usage, cc.argSpecs)
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
		if i < len(cc.argSpecs)-1 {
			if s.Min != 1 || s.Max != 1 {
				panic("only the last argument may take a range of values")
			}
		}
	}
	fn := func(_ *cobra.Command, args []string) error {
		var minCnt, maxCnt int
		c := len(args)
		for _, s := range cc.argSpecs {
			if c < minCnt+s.Min {
				return fmt.Errorf("%s is required", s.Title)
			}
			minCnt += s.Min
			maxCnt = addWithOverflow(maxCnt, s.Max)
		}
		if len(args) > maxCnt {
			return fmt.Errorf("expected at most %d argument(s)", maxCnt)
		}
		return nil
	}
	return fn
}

// addWithOverflow adds two integers and returns the result
// If the sum is greater than math.MaxInt, it returns math.MaxInt.
// a and b are assumed to be non-negative.
func addWithOverflow(a, b int) int {
	if a > math.MaxInt-b {
		return math.MaxInt
	}
	return a + b
}

func makeCommandUsageString(usage string, specs []ArgSpec) string {
	var sb strings.Builder
	sb.WriteString(usage)
	for _, s := range specs {
		sb.WriteByte(' ')
		if s.Min == 0 {
			sb.WriteByte('[')
		} else {
			sb.WriteByte('{')
		}
		sb.WriteString(s.Title)
		if s.Min > 1 {
			for i := 1; i < s.Min; i++ {
				sb.WriteString(", ")
				sb.WriteString(s.Title)
			}
		}
		if s.Max-s.Min > 1 {
			sb.WriteString(", ")
			sb.WriteString("...")
		}
		if s.Min == 0 {
			sb.WriteByte(']')
		} else {
			sb.WriteByte('}')
		}
	}
	sb.WriteString(" [flags]")
	return sb.String()
}
