//go:build std || map

package _map

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type MapLock struct{}

func (mc *MapLock) Unwrappable() {}

func (mc *MapLock) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("lock")
	long := `Lock a key in the given Map

This command is only available in the interactive mode.`
	short := "Lock a key in the given Map"
	cc.SetCommandHelp(long, short)
	addKeyTypeFlag(cc)
	cc.AddIntFlag(mapTTL, "", ttlUnset, false, "time-to-live (ms)")
	cc.AddStringArg(argKey, argTitleKey)
	return nil
}

func (mc *MapLock) Exec(ctx context.Context, ec plug.ExecContext) error {
	mapName := ec.Props().GetString(base.FlagName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		m, err := getMap(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		keyStr := ec.GetStringArg(argKey)
		keyData, err := makeKeyData(ec, ci, keyStr)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Locking key in map %s", mapName))
		if ttl := GetTTL(ec); ttl != ttlUnset {
			return nil, m.LockWithLease(ctx, keyData, time.Duration(GetTTL(ec)))
		}
		return nil, m.Lock(ctx, keyData)
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK Locked the key in map %s", mapName)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:lock", &MapLock{}, plug.OnlyInteractive{}))
}
