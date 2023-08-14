//go:build std || job

package job

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
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
	nameOrID := ec.Args()[0]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Resuming job: %s", nameOrID))
		j := jet.New(ci, sp, ec.Logger())
		jis, err := j.GetJobList(ctx)
		if err != nil {
			return nil, err
		}
		jm, err := NewJobNameToIDMap(jis)
		if err != nil {
			return nil, err
		}
		jid, ok := jm.GetIDForName(nameOrID)
		if !ok {
			return nil, jet.ErrInvalidJobID
		}
		return nil, j.ResumeJob(ctx, jid)
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
		ec.PrintlnUnnecessary(fmt.Sprintf("Job resumed: %s", nameOrID))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job:resume", &ResumeCmd{}))
}
