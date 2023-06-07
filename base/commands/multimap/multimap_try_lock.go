//go:build base || multimap

package _multimap

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client"
)

type MultiMapTryLockCommand struct{}

func (m MultiMapTryLockCommand) Init(cc plug.InitContext) error {
	addKeyTypeFlag(cc)
	cc.AddIntFlag(multiMapTTL, "", ttlUnset, false, "time-to-live (ms)")
	help := "Try to lock a key in the given MultiMap"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("try-lock [key] [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (m MultiMapTryLockCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
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
	req := codec.EncodeMultiMapTryLockRequest(mmName, keyData, 0, 0, ttl, 0)
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Trying to lock multimap %s", mmName))
		return ci.InvokeOnKey(ctx, req, keyData, nil)
	})
	if err != nil {
		return err
	}
	stop()
	resp := codec.DecodeMultiMapTryLockResponse(rv.(*hazelcast.ClientMessage))
	row := output.Row{
		output.Column{
			Name:  output.NameValue,
			Type:  serialization.TypeBool,
			Value: resp,
		},
	}
	if ec.Props().GetBool(multiMapFlagShowType) {
		row = append(row, output.Column{
			Name:  output.NameValueType,
			Type:  serialization.TypeString,
			Value: serialization.TypeToLabel(serialization.TypeBool),
		})
	}
	return ec.AddOutputRows(ctx, row)
}

func init() {
	Must(plug.Registry.RegisterCommand("multimap:try-lock", &MultiMapTryLockCommand{}, true))
}
