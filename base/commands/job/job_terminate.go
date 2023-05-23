package job

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
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
	msg                string
	waitState          int32
}

func (cm TerminateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage(fmt.Sprintf("%s [job-ID/name]", cm.name))
	cc.SetCommandHelp(cm.longHelp, cm.shortHelp)
	cc.SetPositionalArgCount(1, 1)
	cc.AddBoolFlag(flagForce, "", false, false, fmt.Sprintf("force %s the job", cm.name))
	cc.AddBoolFlag(flagWait, "", false, false, "wait for the operation to finish")
	return nil
}

func (cm TerminateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	// just preloading the client
	_, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	tm := cm.terminateMode
	if ec.Props().GetBool(flagForce) {
		tm = cm.terminateModeForce
	}
	jm, err := jet.NewJobNameToIDMap(ctx, ec, false)
	if err != nil {
		return err
	}
	arg := ec.Args()[0]
	jid, ok := jm.GetIDForName(arg)
	if !ok {
		return jet.ErrInvalidJobID
	}
	if err = terminateJob(ctx, ec, jid, arg, tm, cm.msg, cm.waitState); err != nil {
		return err
	}
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	if verbose {
		ec.PrintlnUnnecessary(fmt.Sprintf("Job %sed: %s", cm.name, idToString(jid)))
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job:cancel", &TerminateCmd{
		name:               "cancel",
		longHelp:           "Cancels the job with the given ID or name.",
		shortHelp:          "Cancels the job with the given ID or name",
		terminateMode:      terminateModeCancelGraceful,
		terminateModeForce: terminateModeCancelForceful,
		msg:                "Cancelling the job",
		waitState:          jet.JobStatusFailed,
	}))
	Must(plug.Registry.RegisterCommand("job:suspend", &TerminateCmd{
		name:               "suspend",
		longHelp:           "Suspends the job with the given ID or name.",
		shortHelp:          "Suspends the job with the given ID or name",
		terminateMode:      terminateModeSuspendGraceful,
		terminateModeForce: terminateModeSuspendForceful,
		msg:                "Suspending the job",
		waitState:          jet.JobStatusSuspended,
	}))
	Must(plug.Registry.RegisterCommand("job:restart", &TerminateCmd{
		name:               "restart",
		longHelp:           "Restarts the job with the given ID or name.",
		shortHelp:          "Restarts the job with the given ID or name",
		terminateMode:      terminateModeRestartGraceful,
		terminateModeForce: terminateModeRestartForceful,
		msg:                "Restarting the job",
		waitState:          jet.JobStatusRunning,
	}))
}
