//go:build base || map

package _map

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client"
)

type MapTryLock struct{}

func (mc *MapTryLock) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	long := `Try to lock a key in the given map. Directly returns the result

This command is only available in the interactive mode.`
	short := "Try to lock a key in the given map. Directly returns the result"
	cc.SetCommandHelp(long, short)
	cc.AddIntFlag(mapTTL, "", ttlUnset, false, "time-to-live (ms)")
	cc.SetCommandUsage("try-lock [key] [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (mc *MapTryLock) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(mapFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	mv, err := ec.Props().GetBlocking(mapPropertyName)
	if err != nil {
		return err
	}
	m := mv.(*hazelcast.Map)
	keyStr := ec.Args()[0]
	keyData, err := makeKeyData(ec, ci, keyStr)
	if err != nil {
		return err
	}
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Locking key in map %s", mapName))
		if ttl := GetTTL(ec); ttl != ttlUnset {
			return m.TryLockWithLease(ctx, keyData, time.Duration(GetTTL(ec)))
		}
		return m.TryLock(ctx, keyData)
	})
	if err != nil {
		return err
	}
	stop()
	locked := rv.(bool)
	return ec.AddOutputRows(ctx, output.Row{
		{
			Name:  "Locked",
			Type:  serialization.TypeBool,
			Value: locked,
		},
	})
}

func init() {
	Must(plug.Registry.RegisterCommand("map:try-lock", &MapTryLock{}, plug.OnlyInteractive{}))
}
