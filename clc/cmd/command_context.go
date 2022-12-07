package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
)

type CommandContext struct {
	Cmd           *cobra.Command
	Vpr           *viper.Viper
	stringValues  map[string]*string
	boolValues    map[string]*bool
	intValues     map[string]*int64
	isInteractive bool
	isTopLevel    bool
	groups        map[string]*cobra.Group
}

func NewCommandContext(cmd *cobra.Command, vpr *viper.Viper, isInteractive bool) *CommandContext {
	return &CommandContext{
		Cmd:           cmd,
		Vpr:           vpr,
		stringValues:  map[string]*string{},
		boolValues:    map[string]*bool{},
		intValues:     map[string]*int64{},
		isInteractive: isInteractive,
		groups:        map[string]*cobra.Group{},
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

func (cc *CommandContext) SetPositionalArgCount(min, max int) {
	if min == max {
		cc.Cmd.Args = cobra.ExactArgs(min)
		return
	}
	if min == 0 {
		cc.Cmd.Args = cobra.MaximumNArgs(max)
	}
	if max == 0 {
		cc.Cmd.Args = cobra.MinimumNArgs(min)
	}
	cc.Cmd.Args = cobra.RangeArgs(min, max)
}

func (cc *CommandContext) Hide() {
	cc.Cmd.Hidden = true
}

func (cc *CommandContext) Interactive() bool {
	return cc.isInteractive
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
	cc.groups[id] = &cobra.Group{
		ID:    id,
		Title: title,
	}
}

func (cc *CommandContext) Groups() []*cobra.Group {
	return mk.ValuesOf(cc.groups)
}

func (cc *CommandContext) AddStringConfig(name, value, flag string, help string) {
	cc.Vpr.SetDefault(name, value)
	if flag != "" && !cc.isInteractive {
		f := cc.Cmd.Flag(flag)
		if f != nil {
			Must(cc.Vpr.BindPFlag(name, f))
		}
		return
	}
}

func (cc *CommandContext) SetTopLevel(b bool) {
	cc.isTopLevel = b
}

func (cc *CommandContext) TopLevel() bool {
	return cc.isTopLevel
}
