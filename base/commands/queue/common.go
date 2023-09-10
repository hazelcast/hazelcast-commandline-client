//go:build std || queue

package queue

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func getQueue(ctx context.Context, ec plug.ExecContext, sp clc.Spinner) (*hazelcast.Queue, error) {
	name := ec.Props().GetString(base.FlagName)
	ci, err := cmd.ClientInternal(ctx, ec, sp)
	if err != nil {
		return nil, err
	}
	sp.SetText(fmt.Sprintf("Getting Queue '%s'", name))
	return ci.Client().GetQueue(ctx, name)
}
