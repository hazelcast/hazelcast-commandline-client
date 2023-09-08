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

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
)

func WaitJobState(ctx context.Context, ec plug.ExecContext, sp stage.Statuser[any], jobNameOrID string, state int32, duration time.Duration) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	j := jet.New(ci, sp, ec.Logger())
	for {
		jl, err := j.GetJobList(ctx)
		if err != nil {
			return err
		}
		ok, err := jet.EnsureJobState(jl, jobNameOrID, state)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		ec.Logger().Debugf("Waiting %s for job %s to transition to state %s", duration.String(), jobNameOrID, jet.StatusToString(state))
		time.Sleep(duration)
	}
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

func terminateJob(ctx context.Context, ec plug.ExecContext, tm int32, cm TerminateCmd) error {
	nameOrID := ec.GetStringArg(argJobID)
	stages := []stage.Stage[any]{
		stage.MakeConnectStage[any](ec),
		{
			ProgressMsg: fmt.Sprintf(cm.inProgressMsg, nameOrID),
			SuccessMsg:  fmt.Sprintf(cm.successMsg, nameOrID),
			FailureMsg:  cm.failureMsg,
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
					return nil, fmt.Errorf("%w: %s", jet.ErrInvalidJobID, nameOrID)
				}
				ec.Logger().Info("%s %s (%s)", cm.inProgressMsg, nameOrID, idToString(jid))
				ji, ok := jm.GetInfoForID(jid)
				if !ok {
					return nil, fmt.Errorf("%w: %s", jet.ErrInvalidJobID, nameOrID)
				}
				var coord types.UUID
				if ji.LightJob {
					conns := ci.ConnectionManager().ActiveConnections()
					if len(conns) == 0 {
						return nil, errors.New("not connected")
					}
					coord = conns[0].MemberUUID()
				}
				return nil, j.TerminateJob(ctx, jid, cm.terminateMode, coord)
			},
		},
	}
	wait := ec.Props().GetBool(flagWait)
	if wait {
		stages = append(stages, stage.Stage[any]{
			ProgressMsg: fmt.Sprintf("Waiting for job to be %sed", cm.name),
			SuccessMsg:  fmt.Sprintf("Job %s is %sed", nameOrID, cm.name),
			FailureMsg:  fmt.Sprintf("Failed to %s %s", cm.name, nameOrID),
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				msg := fmt.Sprintf("Waiting for job %s to be %sed", nameOrID, cm.name)
				ec.Logger().Info(msg)
				return nil, WaitJobState(ctx, ec, status, nameOrID, cm.waitState, 1*time.Second)
			},
		})
	}
	_, err := stage.Execute[any](ctx, ec, nil, stage.NewFixedProvider(stages...))
	if err != nil {
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
