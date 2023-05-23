package job

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

func WaitJobState(ctx context.Context, ec plug.ExecContext, msg, jobNameOrID string, state int32, duration time.Duration) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, spinner clc.Spinner) (any, error) {
		if msg != "" {
			spinner.SetText(msg)
		}
		for {
			jl, err := jet.GetJobList(ctx, ci)
			if err != nil {
				return nil, err
			}
			ok, err := jet.EnsureJobState(jl, jobNameOrID, state)
			if err != nil {
				return nil, err
			}
			if ok {
				return nil, nil
			}
			ec.Logger().Debugf("Waiting %s for job %s to transition to state %s", duration.String(), jobNameOrID, jet.StatusToString(state))
			time.Sleep(duration)
		}
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func idToString(id int64) string {
	buf := []byte("0000-0000-0000-0000")
	hex := []byte(strconv.FormatInt(id, 16))
	j := 18
	for i := len(hex) - 1; i >= 0; i-- {
		buf[j] = hex[i]
		if j == 15 || j == 10 || j == 5 {
			j--
		}
		j--
	}
	return string(buf[:])
}

func terminateJob(ctx context.Context, ec plug.ExecContext, jobID int64, nameOrID string, terminateMode int32, text string, waitState int32) error {
	wait := ec.Props().GetBool(flagWait)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("%s %s", text, nameOrID))
		ec.Logger().Info("%s %s (%s)", text, nameOrID, idToString(jobID))
		req := codec.EncodeJetTerminateJobRequest(jobID, terminateMode, types.UUID{})
		if _, err := ci.InvokeOnRandomTarget(ctx, req, nil); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	if wait {
		msg := fmt.Sprintf("Waiting for the operation to finish for job %s", nameOrID)
		ec.Logger().Info(msg)
		err = WaitJobState(ctx, ec, msg, nameOrID, waitState, 1*time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}
