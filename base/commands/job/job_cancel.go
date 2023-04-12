package job

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type CancelCmd struct{}

func (cm CancelCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("cancel [job ID]")
	help := "Cancel a job"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(1, 1)
	cc.AddBoolFlag(flagForce, "", false, false, "force cancel the job")
	return nil
}

func (cm CancelCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	jids := ec.Args()[0]
	jid, err := stringToID(jids)
	if err != nil {
		return fmt.Errorf("invalid job ID: %s: %w", jids, err)
	}
	_, cancel, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Getting the job list")
		tm := terminateModeCancelGraceful
		if ec.Props().GetBool(flagForce) {
			tm = terminateModeCancelForceful
		}
		req := codec.EncodeJetTerminateJobRequest(jid, tm, types.UUID{})
		if _, err := ci.InvokeOnRandomTarget(ctx, req, nil); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	defer cancel()
	if verbose {
		I2(fmt.Fprintf(ec.Stderr(), "Job cancellation started: %s\n", idToString(jid)))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job:cancel", &CancelCmd{}))
}
