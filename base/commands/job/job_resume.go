package job

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type ResumeCmd struct{}

func (cm ResumeCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("resume [job-ID/name]")
	help := "Resumes a suspended job"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(1, 1)
	cc.AddBoolFlag(flagWait, "", false, false, "wait for the job to be resumed")
	return nil
}

func (cm ResumeCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	jm, err := jet.NewJobNameToIDMap(ctx, ec, false)
	if err != nil {
		return err
	}
	nameOrID := ec.Args()[0]
	jid, ok := jm.GetIDForName(nameOrID)
	if !ok {
		return jet.ErrInvalidJobID
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Resuming job: %s", idToString(jid)))
		req := codec.EncodeJetResumeJobRequest(jid)
		if _, err := ci.InvokeOnRandomTarget(ctx, req, nil); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	if ec.Props().GetBool(flagWait) {
		msg := fmt.Sprintf("Waiting for job %s to start", nameOrID)
		ec.Logger().Info(msg)
		err = WaitJobState(ctx, ec, msg, nameOrID, jet.JobStatusRunning, 2*time.Second)
		if err != nil {
			return err
		}
	}
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	if verbose {
		ec.PrintlnUnnecessary(fmt.Sprintf("Job resumed: %s", idToString(jid)))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job:resume", &ResumeCmd{}))
}
