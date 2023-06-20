//go:build base || list

package list

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type ListRemoveIndexCommand struct{}

func (mc *ListRemoveIndexCommand) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(1, 1)
	help := "Remove the value at the given index in the list"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("remove-index [index] [flags]")
	return nil
}

func (mc *ListRemoveIndexCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Props().GetString(listFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	// get the list just to ensure the corresponding proxy is created
	if _, err := ec.Props().GetBlocking(listPropertyName); err != nil {
		return err
	}
	index, err := strconv.Atoi(ec.Args()[0])
	if err != nil {
		return err
	}
	if index < 0 {
		return errors.New("index cannot be smaller than 0")
	}
	pid, err := stringToPartitionID(ci, name)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Removing value from the list %s", name))
		req := codec.EncodeListRemoveWithIndexRequest(name, int32(index))
		return ci.InvokeOnPartition(ctx, req, pid, nil)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("list:remove-index", &ListRemoveIndexCommand{}))
}
