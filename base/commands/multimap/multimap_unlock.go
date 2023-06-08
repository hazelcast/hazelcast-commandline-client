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

type MultiMapUnlockCommand struct{}

func (m MultiMapUnlockCommand) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	help := "Unlock a key in the given MultiMap"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("unlock [key] [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (m MultiMapUnlockCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mmName := ec.Props().GetString(multiMapFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	keyData, err := makeKeyData(ec, ci, keyStr)
	if err != nil {
		return err
	}
	req := codec.EncodeMultiMapUnlockRequest(mmName, keyData, 0, 0)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Unlocking key of multimap %s", mmName))
		return ci.InvokeOnKey(ctx, req, keyData, nil)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("multi-map:unlock", &MultiMapUnlockCommand{}, plug.OnlyInteractive{}))
}
