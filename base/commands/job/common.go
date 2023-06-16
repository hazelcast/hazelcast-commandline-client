package job

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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

func terminateJob(ctx context.Context, ec plug.ExecContext, jobID int64, nameOrID string, terminateMode int32, text string, waitState int32) error {
	wait := ec.Props().GetBool(flagWait)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("%s %s", text, nameOrID))
		ec.Logger().Info("%s %s (%s)", text, nameOrID, idToString(jobID))
		j := jet.New(ci, sp, ec.Logger())
		return nil, j.TerminateJob(ctx, jobID, terminateMode)
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

func MakeJobNameIDMaps(jobList []control.JobAndSqlSummary) (jobNameToID map[string]int64, idToJobName map[int64]string) {
	jobNameToID = make(map[string]int64, len(jobList))
	idToJobName = make(map[int64]string, len(jobList))
	for _, j := range jobList {
		idToJobName[j.JobId] = j.NameOrId
		if j.Status == jet.JobStatusFailed || j.Status == jet.JobStatusCompleted {
			continue
		}
		jobNameToID[j.NameOrId] = j.JobId

	}
	return jobNameToID, idToJobName
}

type JobNameToIDMap struct {
	nameToID map[string]int64
	IDToName map[int64]string
}

func NewJobNameToIDMap(ctx context.Context, ec plug.ExecContext, forceLoadJobList bool) (*JobNameToIDMap, error) {
	hasJobName := false
	for _, arg := range ec.Args() {
		if _, err := stringToID(arg); err != nil {
			hasJobName = true
			break
		}
	}
	if !hasJobName && !forceLoadJobList {
		// relies on m.GetIDForName returning the numeric jobID
		// if s is a UUID
		return &JobNameToIDMap{
			nameToID: map[string]int64{},
			IDToName: map[int64]string{},
		}, nil
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, err
	}
	jl, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Getting job list")
		j := jet.New(ci, sp, ec.Logger())
		return j.GetJobList(ctx)
	})
	if err != nil {
		return nil, err
	}
	stop()
	n2i, i2j := MakeJobNameIDMaps(jl.([]control.JobAndSqlSummary))
	return &JobNameToIDMap{
		nameToID: n2i,
		IDToName: i2j,
	}, nil
}

func (m JobNameToIDMap) GetIDForName(idOrName string) (int64, bool) {
	id, err := stringToID(idOrName)
	// note that comparing err to nil
	if err == nil {
		return id, true
	}
	v, ok := m.nameToID[idOrName]
	return v, ok
}

func (m JobNameToIDMap) GetNameForID(id int64) (string, bool) {
	v, ok := m.IDToName[id]
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
