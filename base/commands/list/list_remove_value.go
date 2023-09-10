//go:build std || list

package list

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ListRemoveValueCommand struct{}

func (mc *ListRemoveValueCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("remove-value")
	help := "Remove a value from the given list"
	cc.SetCommandHelp(help, help)
	addValueTypeFlag(cc)
	cc.AddStringArg(base.ArgValue, base.ArgTitleValue)
	return nil
}

func (mc *ListRemoveValueCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	value := ec.GetStringArg(base.ArgValue)
	return removeFromList(ctx, ec, name, 0, value)
}

func init() {
	Must(plug.Registry.RegisterCommand("list:remove-value", &ListRemoveValueCommand{}))
}
