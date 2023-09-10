//go:build std || job

package job

import (
	"context"
	"fmt"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type TerminateCmd struct {
	name               string
	longHelp           string
	shortHelp          string
	terminateMode      int32
	terminateModeForce int32
	inProgressMsg      string
	successMsg         string
	failureMsg         string
	waitState          int32
}

func (cm TerminateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage(cm.name)
	cc.SetCommandHelp(cm.longHelp, cm.shortHelp)
	cc.AddBoolFlag(flagForce, "", false, false, fmt.Sprintf("force %s the job", cm.name))
	cc.AddBoolFlag(flagWait, "", false, false, "wait for the operation to finish")
	cc.AddStringArg(argJobID, argTitleJobID)
	return nil
}

func (cm TerminateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	tm := cm.terminateMode
	if ec.Props().GetBool(flagForce) {
		tm = cm.terminateModeForce
	}
	if err := terminateJob(ctx, ec, tm, cm); err != nil {
		return err
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job:cancel", &TerminateCmd{
		name:               "cancel",
		longHelp:           "Cancels the job with the given ID or name",
		shortHelp:          "Cancels the job with the given ID or name",
		terminateMode:      jet.TerminateModeCancelGraceful,
		terminateModeForce: jet.TerminateModeCancelForceful,
		waitState:          jet.JobStatusFailed,
		inProgressMsg:      "Starting to cancel %s",
		successMsg:         "Started cancellation of %s",
		failureMsg:         "Failed to start job cancellation",
	}))
	Must(plug.Registry.RegisterCommand("job:suspend", &TerminateCmd{
		name:               "suspend",
		longHelp:           "Suspends the job with the given ID or name",
		shortHelp:          "Suspends the job with the given ID or name",
		terminateMode:      jet.TerminateModeSuspendGraceful,
		terminateModeForce: jet.TerminateModeSuspendForceful,
		waitState:          jet.JobStatusSuspended,
		inProgressMsg:      "Starting to suspend %s",
		successMsg:         "Started suspension of %s",
		failureMsg:         "Failed to start job suspension",
	}))
	Must(plug.Registry.RegisterCommand("job:restart", &TerminateCmd{
		name:               "restart",
		longHelp:           "Restarts the job with the given ID or name",
		shortHelp:          "Restarts the job with the given ID or name",
		terminateMode:      jet.TerminateModeRestartGraceful,
		terminateModeForce: jet.TerminateModeRestartForceful,
		waitState:          jet.JobStatusRunning,
		inProgressMsg:      "Initiating the restart of %s",
		successMsg:         "Initiated the restart of %s",
		failureMsg:         "Failed to initiate job restart",
	}))
}
