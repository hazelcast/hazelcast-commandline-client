//go:build std || job

package job

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
)

func WaitJobState(ctx context.Context, ec plug.ExecContext, msg, jobNameOrID string, state int32, duration time.Duration) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		if msg != "" {
			sp.SetText(msg)
		}
		j := jet.New(ci, sp, ec.Logger())
		for {
			jl, err := j.GetJobList(ctx)
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

func terminateJob(ctx context.Context, ec plug.ExecContext, name string, terminateMode int32, text string, waitState int32) error {
	nameOrID := ec.GetStringArg(argJobID)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	jidv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("%s %s", text, nameOrID))
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
			return nil, fmt.Errorf("%w: %s", jet.ErrInvalidJobID, nameOrID)
		}
		ec.Logger().Info("%s %s (%s)", text, nameOrID, idToString(jid))
		ji, ok := jm.GetInfoForID(jid)
		if !ok {
			return nil, jet.ErrInvalidJobID
		}
		var coord types.UUID
		if ji.LightJob {
			conns := ci.ConnectionManager().ActiveConnections()
			if len(conns) == 0 {
				return nil, errors.New("not connected")
			}
			coord = conns[0].MemberUUID()
		}
		return jid, j.TerminateJob(ctx, jid, terminateMode, coord)
	})
	if err != nil {
		return err
	}
	stop()
	err = nil
	if ec.Props().GetBool(flagWait) {
		msg := fmt.Sprintf("Waiting for the operation to finish for job %s", nameOrID)
		ec.Logger().Info(msg)
		err = WaitJobState(ctx, ec, msg, nameOrID, waitState, 1*time.Second)
	}
	if err != nil {
		if ec.Props().GetBool(clc.PropertyVerbose) {
			ec.PrintlnUnnecessary(fmt.Sprintf("Job %sed: %s", name, idToString(jidv.(int64))))
		}
		return err
	}
	return nil
}

func MakeJobNameIDMaps(jobList []control.JobAndSqlSummary) (jobNameToID map[string]int64, jobIDToInfo map[int64]control.JobAndSqlSummary) {
	jobNameToID = make(map[string]int64, len(jobList))
	jobIDToInfo = make(map[int64]control.JobAndSqlSummary, len(jobList))
	for _, j := range jobList {
		jobIDToInfo[j.JobId] = j
		if j.Status == jet.JobStatusFailed || j.Status == jet.JobStatusCompleted {
			continue
		}
		jobNameToID[j.NameOrId] = j.JobId

	}
	return jobNameToID, jobIDToInfo
}

type JobsInfo struct {
	nameToID map[string]int64
	IDToInfo map[int64]control.JobAndSqlSummary
}

func NewJobNameToIDMap(jobs []control.JobAndSqlSummary) (*JobsInfo, error) {
	n2i, i2j := MakeJobNameIDMaps(jobs)
	return &JobsInfo{
		nameToID: n2i,
		IDToInfo: i2j,
	}, nil
}

func (m JobsInfo) GetIDForName(idOrName string) (int64, bool) {
	id, err := stringToID(idOrName)
	// note that comparing err to nil
	if err == nil {
		return id, true
	}
	v, ok := m.nameToID[idOrName]
	return v, ok
}

func (m JobsInfo) GetInfoForID(id int64) (control.JobAndSqlSummary, bool) {
	v, ok := m.IDToInfo[id]
	return v, ok
}

func stringToID(s string) (int64, error) {
	// first try whether it's an int
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// otherwise this can be an ID
		s = strings.Replace(s, "-", "", -1)
		i, err = strconv.ParseInt(s, 16, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid ID: %s: %w", s, err)
		}
	}
	return i, nil
}
