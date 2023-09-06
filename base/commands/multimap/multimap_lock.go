//go:build std || multimap

package multimap

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type MultiMapLockCommand struct{}

func (m MultiMapLockCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("lock")
	long := `Lock a key in the given MultiMap

This command is only available in the interactive mode.`
	short := "Lock a key in the given MultiMap"
	cc.SetCommandHelp(long, short)
	addKeyTypeFlag(cc)
	cc.AddStringArg(argKey, argTitleKey)
	return nil
}

func (m MultiMapLockCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mmName := ec.Props().GetString(multiMapFlagName)
	mv, err := ec.Props().GetBlocking(multiMapPropertyName)
	if err != nil {
		return err
	}
	keyStr := ec.GetStringArg(argKey)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	keyData, err := makeKeyData(ec, ci, keyStr)
	if err != nil {
		return err
	}
	mm := mv.(*hazelcast.MultiMap)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Locking key of multimap %s", mmName))
		if ttl := GetTTL(ec); ttl != ttlUnset {
			return mm.LockWithLease(ctx, keyData, time.Duration(GetTTL(ec))), nil
		}
		return mm.Lock(ctx, keyData), nil
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("multi-map:lock", &MultiMapLockCommand{}, plug.OnlyInteractive{}))
}
