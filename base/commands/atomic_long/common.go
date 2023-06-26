//go:build base || atomiclong

package atomiclong

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func atomicLongChangeValue(ctx context.Context, ec plug.ExecContext, verb string, change func(int64) int64) error {
	al, err := ec.Props().GetBlocking(atomicLongPropertyName)
	if err != nil {
		return err
	}
	by := ec.Props().GetInt(atomicLongFlagBy)
	by = change(by)
	ali := al.(*hazelcast.AtomicLong)
	vali, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("%sing the AtomicLong %s", verb, ali.Name()))
		val, err := ali.AddAndGet(ctx, by)
		if err != nil {
			return nil, err
		}
		return val, nil
	})
	if err != nil {
		return err
	}
	stop()
	val := vali.(int64)
	row := output.Row{
		output.Column{
			Name:  "Value",
			Type:  serialization.TypeInt64,
			Value: val,
		},
	}
	return ec.AddOutputRows(ctx, row)

}
