//go:build std || job

package job

import (
	"context"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ResumeCmd struct{}

func (cm ResumeCmd) Unwrappable() {}

func (cm ResumeCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("resume")
	help := "Resumes a suspended job"
	cc.SetCommandHelp(help, help)
	cc.AddBoolFlag(flagWait, "", false, false, "wait for the job to be resumed")
	cc.AddStringArg(argJobID, argTitleJobID)
	return nil
}

func (cm ResumeCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	nameOrID := ec.GetStringArg(argJobID)
	stages := []stage.Stage[any]{
		makeConnectStage(ec),
		{
			ProgressMsg: fmt.Sprintf("Initiating resume of job: %s", nameOrID),
			SuccessMsg:  fmt.Sprintf("Initiated resume of job %s", nameOrID),
			FailureMsg:  fmt.Sprintf("Failed initiating job resume %s", nameOrID),
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				ci, err := ec.ClientInternal(ctx)
				if err != nil {
					return nil, err
				}
				j := jet.New(ci, status, ec.Logger())
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
			},
		},
	}
	if ec.Props().GetBool(flagWait) {
		stages = append(stages, stage.Stage[any]{
			ProgressMsg: fmt.Sprintf("Waiting for job %s to resume", nameOrID),
			SuccessMsg:  fmt.Sprintf("Job %s is resumed", nameOrID),
			FailureMsg:  fmt.Sprintf("Job %s failed to resume", nameOrID),
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				return nil, WaitJobState(ctx, ec, status, nameOrID, jet.JobStatusRunning, 2*time.Second)
			},
		})
	}
	_, err := stage.Execute[any](ctx, ec, nil, stage.NewFixedProvider(stages...))
	if err != nil {
		return err
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job:resume", &ResumeCmd{}))
}
