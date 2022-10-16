package internal

import (
	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/make"
)

type CommandContext struct {
	Cmd           *cobra.Command
	stringValues  map[string]*string
	boolValues    map[string]*bool
	isInteractive bool
	groups        map[string]*cobra.Group
}

func NewCommandContext(cmd *cobra.Command, isInteractive bool) *CommandContext {
	return &CommandContext{
		Cmd:           cmd,
		stringValues:  map[string]*string{},
		boolValues:    map[string]*bool{},
		isInteractive: isInteractive,
		groups:        map[string]*cobra.Group{},
	}
}

func (cc *CommandContext) AddStringFlag(long, short, value string, required bool, help string) {
	var s string
	if short != "" {
		cc.Cmd.PersistentFlags().StringVarP(&s, long, short, value, help)
	} else {
		cc.Cmd.PersistentFlags().StringVar(&s, long, value, help)
	}
	if required {
		check.Must(cc.Cmd.MarkPersistentFlagRequired(long))
	}
	cc.stringValues[long] = &s
}

func (cc *CommandContext) AddBoolFlag(long, short string, value bool, required bool, help string) {
	cc.Cmd.AddGroup()
	var b bool
	if short != "" {
		cc.Cmd.PersistentFlags().BoolVarP(&b, long, short, value, help)
	} else {
		cc.Cmd.PersistentFlags().BoolVar(&b, long, value, help)
	}
	if required {
		check.Must(cc.Cmd.MarkPersistentFlagRequired(long))
	}
	cc.boolValues[long] = &b
}

func (cc *CommandContext) Interactive() bool {
	return cc.isInteractive
}

func (cc *CommandContext) SetCommandUsage(long, short string) {
	if long != "" {
		cc.Cmd.Long = long
	}
	if short != "" {
		cc.Cmd.Short = short
	}
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
	return make.SliceFromMap(cc.groups)
}
