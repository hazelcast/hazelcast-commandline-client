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

type ResumeCommand struct{}

func (ResumeCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("resume")
	help := "Resumes a suspended job"
	cc.SetCommandHelp(help, help)
	cc.AddBoolFlag(flagWait, "", false, false, "wait for the job to be resumed")
	cc.AddStringArg(argJobID, argTitleJobID)
	return nil
}

func (ResumeCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	nameOrID := ec.GetStringArg(argJobID)
	stages := []stage.Stage[int64]{
		stage.MakeConnectStage[int64](ec),
		{
			ProgressMsg: fmt.Sprintf("Initiating resume of job: %s", nameOrID),
			SuccessMsg:  fmt.Sprintf("Initiated resume of job %s", nameOrID),
			FailureMsg:  fmt.Sprintf("Failed initiating job resume %s", nameOrID),
			Func: func(ctx context.Context, status stage.Statuser[int64]) (int64, error) {
				ci, err := ec.ClientInternal(ctx)
				if err != nil {
					return 0, err
				}
				j := jet.New(ci, status, ec.Logger())
				jis, err := j.GetJobList(ctx)
				if err != nil {
					return 0, err
				}
				jm, err := NewJobNameToIDMap(jis)
				if err != nil {
					return 0, err
				}
				jid, ok := jm.GetIDForName(nameOrID)
				if !ok {
					return 0, jet.ErrInvalidJobID
				}
				return jid, j.ResumeJob(ctx, jid)
			},
		},
	}
	if ec.Props().GetBool(flagWait) {
		stages = append(stages, stage.Stage[int64]{
			ProgressMsg: fmt.Sprintf("Waiting for job %s to resume", nameOrID),
			SuccessMsg:  fmt.Sprintf("Job %s is resumed", nameOrID),
			FailureMsg:  fmt.Sprintf("Job %s failed to resume", nameOrID),
			Func: func(ctx context.Context, status stage.Statuser[int64]) (int64, error) {
				jobID := status.Value()
				return jobID, WaitJobState(ctx, ec, status, jet.JobStatusRunning, 2*time.Second)
			},
		})
	}
	_, err := stage.Execute(ctx, ec, 0, stage.NewFixedProvider(stages...))
	if err != nil {
		return err
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job:resume", &ResumeCommand{}))
}
