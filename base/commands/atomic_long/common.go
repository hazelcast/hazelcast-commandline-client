//go:build std || atomiclong

package atomiclong

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func atomicLongChangeValue(ctx context.Context, ec plug.ExecContext, verb string, change func(int64) int64) error {
	by := ec.Props().GetInt(atomicLongFlagBy)
	val, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ali, err := getAtomicLong(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("%sing the AtomicLong %s", verb, ali.Name()))
		val, err := ali.AddAndGet(ctx, change(by))
		if err != nil {
			return nil, err
		}
		msg := fmt.Sprintf("OK %sed AtomicLong %s by %d.\n", verb, ali.Name(), by)
		ec.PrintlnUnnecessary(msg)
		return val, nil
	})
	if err != nil {
		return err
	}
	stop()
	row := output.Row{
		output.Column{
			Name:  "Value",
			Type:  serialization.TypeInt64,
			Value: val,
		},
	}
	return ec.AddOutputRows(ctx, row)
}

func getAtomicLong(ctx context.Context, ec plug.ExecContext, sp clc.Spinner) (*hazelcast.AtomicLong, error) {
	name := ec.Props().GetString(flagName)
	ci, err := cmd.ClientInternal(ctx, ec, sp)
	if err != nil {
		return nil, err
	}
	sp.SetText(fmt.Sprintf("Getting atomic long %s", name))
	return ci.Client().CPSubsystem().GetAtomicLong(ctx, name)
}
