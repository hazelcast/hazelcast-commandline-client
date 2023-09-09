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

type executeState struct {
	Name  string
	Value int64
}

func atomicLongChangeValue(ctx context.Context, ec plug.ExecContext, verb string, change func(int64) int64) error {
	by := ec.Props().GetInt(atomicLongFlagBy)
	stateV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ali, err := getAtomicLong(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("%sing the AtomicLong %s", verb, ali.Name()))
		val, err := ali.AddAndGet(ctx, change(by))
		if err != nil {
			return nil, err
		}
		state := executeState{
			Name:  ali.Name(),
			Value: val,
		}
		return state, nil
	})
	if err != nil {
		return err
	}
	stop()
	s := stateV.(executeState)
	msg := fmt.Sprintf("OK %sed AtomicLong %s by %d.\n", verb, s.Name, by)
	ec.PrintlnUnnecessary(msg)
	row := output.Row{
		output.Column{
			Name:  "Value",
			Type:  serialization.TypeInt64,
			Value: s.Value,
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
	sp.SetText(fmt.Sprintf("Getting AtomicLong '%s'", name))
	return ci.Client().CPSubsystem().GetAtomicLong(ctx, name)
}
