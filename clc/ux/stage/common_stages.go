package stage

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func MakeConnectStage[T any](ec plug.ExecContext) Stage[T] {
	s := Stage[T]{
		ProgressMsg: "Connecting to the cluster",
		SuccessMsg:  "Connected to the cluster",
		FailureMsg:  "Failed connecting to the cluster",
		Func: func(ctx context.Context, status Statuser[T]) (T, error) {
			var v T
			_, err := ec.ClientInternal(ctx)
			if err != nil {
				return v, err
			}
			return v, nil
		},
	}
	return s
}
