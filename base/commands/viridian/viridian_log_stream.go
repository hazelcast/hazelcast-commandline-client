package viridian

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type StreamLogCmd struct{}

func (cm StreamLogCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("stream-logs [cluster-ID/name]")
	long := `Streams logs of the given Viridian cluster.

Make sure you login before running this command.
`
	short := "Streams logs of a Viridian cluster"
	cc.SetCommandHelp(long, short)
}

func (cm StreamLogCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	//TODO implement me
	panic("implement me")
}
