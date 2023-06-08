//go:build base || multimap

package _multimap

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
	mv, err := ec.Props().GetBlocking(multiMapPropertyName)
	if err != nil {
		return err
	}
	keyStr := ec.Args()[0]
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	keyData, err := makeKeyData(ec, ci, keyStr)
	if err != nil {
		return err
	}
	mm := mv.(*hazelcast.MultiMap)
	lv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Trying to lock multimap %s", mmName))
		if ttl := GetTTL(ec); ttl != ttlUnset {
			return mm.TryLockWithLease(ctx, keyData, time.Duration(GetTTL(ec)))
		}
		return mm.TryLock(ctx, keyData)
	})
	if err != nil {
		return err
	}
	stop()
	row := output.Row{
		output.Column{
			Name:  output.NameValue,
			Type:  serialization.TypeBool,
			Value: lv.(bool),
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
	Must(plug.Registry.RegisterCommand("multi-map:try-lock", &MultiMapTryLockCommand{}, plug.OnlyInteractive{}))
}
