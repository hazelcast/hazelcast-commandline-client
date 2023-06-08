//go:build base || multimap

package _multimap

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
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
	ttl := GetTTL(ec)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	keyData, err := makeKeyData(ec, ci, keyStr)
	if err != nil {
		return err
	}
	req := codec.EncodeMultiMapLockRequest(mmName, keyData, 0, ttl, 0)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Locking key of multimap %s", mmName))
		return ci.InvokeOnKey(ctx, req, keyData, nil)
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
