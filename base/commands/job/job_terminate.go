package job

import (
	"context"
	"fmt"
	"math"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type TerminateCmd struct {
	name               string
	longHelp           string
	shortHelp          string
	terminateMode      int32
	terminateModeForce int32
	msg                string
}

func (cm TerminateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage(fmt.Sprintf("%s [job ID/name, ...]", cm.name))
	cc.SetCommandHelp(cm.longHelp, cm.shortHelp)
	cc.SetPositionalArgCount(1, math.MaxInt)
	cc.AddBoolFlag(flagForce, "", false, false, fmt.Sprintf("force %s the job", cm.name))
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
	var allErrs []error
	jm, err := newJobNameToIDMap(ctx, ec, false)
	if err != nil {
		return err
	}
	for _, arg := range ec.Args() {
		jid, ok := jm.GetIDForName(arg)
		if !ok {
			allErrs = append(allErrs, errInvalidJobID)
			continue
		}
		if err := terminateJob(ctx, ec, jid, tm, cm.msg); err != nil {
			allErrs = append(allErrs, fmt.Errorf("%s: %w", arg, err))
		} else {
			verbose := ec.Props().GetBool(clc.PropertyVerbose)
			if verbose {
				I2(fmt.Fprintf(ec.Stderr(), "Job %s: %s\n", cm.name, idToString(jid)))
			}
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return makeErrorsString(allErrs)
}

func init() {
	Must(plug.Registry.RegisterCommand("job:cancel", &TerminateCmd{
		name:               "cancel",
		longHelp:           "Cancels the job with the given ID or name",
		shortHelp:          "Cancels the job with the given ID or name",
		terminateMode:      terminateModeCancelGraceful,
		terminateModeForce: terminateModeCancelForceful,
		msg:                "Cancelling the job",
	}))
	Must(plug.Registry.RegisterCommand("job:suspend", &TerminateCmd{
		name:               "suspend",
		longHelp:           "Suspends the job with the given ID or name",
		shortHelp:          "Suspends the job with the given ID or name",
		terminateMode:      terminateModeSuspendGraceful,
		terminateModeForce: terminateModeSuspendForceful,
		msg:                "Suspending the job",
	}))
	Must(plug.Registry.RegisterCommand("job:restart", &TerminateCmd{
		name:               "restart",
		longHelp:           "Restarts the job with the given ID or name",
		shortHelp:          "Restarts the job with the given ID or name",
		terminateMode:      terminateModeRestartGraceful,
		terminateModeForce: terminateModeRestartForceful,
		msg:                "Restarting the job",
	}))
}
