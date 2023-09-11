//go:build std || atomiclong

package atomiclong

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type executeState struct {
	Name  string
	Value int64
}

func atomicLongChangeValue(ctx context.Context, ec plug.ExecContext, verb string, change func(int64) int64) error {
	name := ec.Props().GetString(base.FlagName)
	by := ec.Props().GetInt(flagBy)
	row, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (output.Row, error) {
		ali, err := getAtomicLong(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("%sing the AtomicLong %s", verb, name))
		val, err := ali.AddAndGet(ctx, change(by))
		if err != nil {
			return nil, err
		}
		s := executeState{
			Name:  name,
			Value: val,
		}
		row := output.Row{
			output.Column{
				Name:  "Value",
				Type:  serialization.TypeInt64,
				Value: s.Value,
			},
		}
		return row, nil
	})
	if err != nil {
		return err
	}
	stop()
	msg := fmt.Sprintf("OK %sed AtomicLong %s by %d.\n", verb, name, by)
	ec.PrintlnUnnecessary(msg)
	return ec.AddOutputRows(ctx, row)
}

func getAtomicLong(ctx context.Context, ec plug.ExecContext, sp clc.Spinner) (*hazelcast.AtomicLong, error) {
	name := ec.Props().GetString(base.FlagName)
	ci, err := cmd.ClientInternal(ctx, ec, sp)
	if err != nil {
		return nil, err
	}
	sp.SetText(fmt.Sprintf("Getting AtomicLong '%s'", name))
	return ci.Client().CPSubsystem().GetAtomicLong(ctx, name)
}
