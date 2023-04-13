package job

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
)

var errInvalidJobID = errors.New("invalid job ID")

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

func terminateJob(ctx context.Context, ec plug.ExecContext, jobID int64, terminateMode int32, text string) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("%s: %s", text, jobID))
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
	return nil
}

func makeErrorsString(errs []error) error {
	var sb strings.Builder
	for _, e := range errs {
		sb.WriteString(fmt.Sprintf("- %s\n", e.Error()))
	}
	return errors.New(sb.String())
}

func getJobList(ctx context.Context, ci *hazelcast.ClientInternal) ([]control.JobAndSqlSummary, error) {
	req := codec.EncodeJetGetJobAndSqlSummaryListRequest()
	resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	ls := codec.DecodeJetGetJobAndSqlSummaryListResponse(resp)
	return ls, nil
}

func makeJobNameToIDMap(jobList []control.JobAndSqlSummary) map[string]int64 {
	m := make(map[string]int64, len(jobList))
	for _, j := range jobList {
		if j.Status == statusFailed || j.Status == statusCompleted {
			continue
		}
		m[j.NameOrId] = j.JobId
	}
	return m
}

type jobNameToIDMap map[string]int64

func newJobNameToIDMap(ctx context.Context, ec plug.ExecContext) (jobNameToIDMap, error) {
	hasJobName := false
	for _, arg := range ec.Args() {
		if _, err := stringToID(arg); err != nil {
			hasJobName = true
			break
		}
	}
	if !hasJobName {
		// relies on m.Get returning the numeric jobID
		// if s is a UUID
		return jobNameToIDMap{}, nil
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, err
	}
	jl, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Getting job list")
		return getJobList(ctx, ci)
	})
	if err != nil {
		return nil, err
	}
	stop()
	m := makeJobNameToIDMap(jl.([]control.JobAndSqlSummary))
	return jobNameToIDMap(m), nil
}

func (m jobNameToIDMap) Get(idOrName string) (int64, bool) {
	id, err := stringToID(idOrName)
	// note that comparing err to nil
	if err == nil {
		return id, true
	}
	v, ok := m[idOrName]
	return v, ok
}
