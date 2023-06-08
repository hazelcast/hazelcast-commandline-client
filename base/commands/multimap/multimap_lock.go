//go:build base || multimap

package _multimap

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

type MultiMapLockCommand struct{}

func (m MultiMapLockCommand) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	cc.AddIntFlag(multiMapTTL, "", ttlUnset, false, "time-to-live (ms)")
	help := "Lock a key in the given MultiMap"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("lock [key] [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (m MultiMapLockCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mmName := ec.Props().GetString(multiMapFlagName)
	mv, err := ec.Props().GetBlocking(multiMapPropertyName)
	if err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	mm := mv.(*hazelcast.MultiMap)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Locking key of multimap %s", mmName))
		if ttl := GetTTL(ec); ttl != ttlUnset {
			return mm.LockWithLease(ctx, keyStr, time.Duration(GetTTL(ec))), nil
		}
		return mm.Lock(ctx, keyStr), nil
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("multimap:lock", &MultiMapLockCommand{}, true))
}
