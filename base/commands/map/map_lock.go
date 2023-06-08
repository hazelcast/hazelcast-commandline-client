//go:build base || map

package _map

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

type MapLock struct{}

func (mc *MapLock) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	long := `Lock a key in the given Map

This command is only available in the interactive mode.`
	short := "Lock a key in the given Map"
	cc.SetCommandHelp(long, short)
	cc.AddIntFlag(mapTTL, "", ttlUnset, false, "time-to-live (ms)")
	cc.SetCommandUsage("lock [key] [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (mc *MapLock) Exec(ctx context.Context, ec plug.ExecContext) error {
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
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
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
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:lock", &MapLock{}, plug.OnlyInteractive{}))
}

const (
	defaultLockID = 0
)

type lockID int64
type lockIDKey struct{}

// extractLockID extracts lock ID from the context.
// If the lock ID is not found, it returns the default lock ID.
func extractLockID(ctx context.Context) int64 {
	if ctx == nil {
		return defaultLockID
	}
	lidv := ctx.Value(lockIDKey{})
	if lidv == nil {
		return defaultLockID
	}
	lid, ok := lidv.(lockID)
	if !ok {
		return defaultLockID
	}
	return int64(lid)
}
