package _set

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-go-client"
)

type SetRemoveCommand struct{}

func (sc *SetRemoveCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	cc.SetPositionalArgCount(1, math.MaxInt)
	help := "Remove values to the given Set"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("remove [values] [flags]")
	return nil
}

func (sc *SetRemoveCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	setName := ec.Props().GetString(setFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	for _, arg := range ec.Args() {
		vd, err := makeValueData(ec, ci, arg)
		if err != nil {
			return err
		}
		req := codec.EncodeSetRemoveRequest(setName, vd)
		pID, err := stringToPartitionID(ci, setName)
		if err != nil {
			return err
		}
		sv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Removing from set %s", setName))
			return ci.InvokeOnPartition(ctx, req, pID, nil)
		})
		if err != nil {
			return err
		}
		stop()
		done := codec.DecodeSetRemoveResponse(sv.(*hazelcast.ClientMessage))
		if !done {
			err = fmt.Errorf("the value for %s was not decoded, due to an unknown error", arg)
			ec.Logger().Error(err)
			return err
		}
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("set:remove", &SetRemoveCommand{}))
}

func stringToPartitionID(ci *hazelcast.ClientInternal, name string) (int32, error) {
	idx := strings.Index(name, "@")
	if keyData, err := ci.EncodeData(name[idx+1:]); err != nil {
		return 0, err
	} else if partitionID, err := ci.GetPartitionID(keyData); err != nil {
		return 0, err
	} else {
		return partitionID, nil
	}
}
