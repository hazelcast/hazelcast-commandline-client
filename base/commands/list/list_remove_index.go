//go:build std || list

package list

import (
	"context"
	"errors"
	"math"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ListRemoveIndexCommand struct{}

func (mc *ListRemoveIndexCommand) Unwrappable() {}

func (mc *ListRemoveIndexCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("remove-index")
	help := "Remove the value at the given index in the list"
	cc.SetCommandHelp(help, help)
	cc.AddInt64Arg(argIndex, argTitleIndex)
	return nil
}

func (mc *ListRemoveIndexCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(base.FlagName)
	index := ec.GetInt64Arg(argIndex)
	if index < 0 {
		return errors.New("index must be non-negative")
	}
	if index > math.MaxInt32 {
		return errors.New("index must fit into a 32bit unsigned integer")
	}
	return removeFromList(ctx, ec, name, int32(index), "")
}

func init() {
	Must(plug.Registry.RegisterCommand("list:remove-index", &ListRemoveIndexCommand{}))
}
