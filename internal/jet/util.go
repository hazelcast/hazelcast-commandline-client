package jet

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
)

const (
	defaultBatchSize                    = 2 * 1024 * 1024 // 10MB
	JobStatusNotRunning                 = 0
	JobStatusStarting                   = 1
	JobStatusRunning                    = 2
	JobStatusSuspended                  = 3
	JobStatusSuspendedExportingSnapshot = 4
	JobStatusCompleting                 = 5
	JobStatusFailed                     = 6
	JobStatusCompleted                  = 7
)

func hashOfPath(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return calculateHash(f)
}

func calculateHash(r io.Reader) ([]byte, error) {
	h := sha256.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, 32)
	b := h.Sum(buf)
	return b, nil
}

func partCountOf(path string, partSize int) (int, error) {
	st, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return int(math.Ceil(float64(st.Size()) / float64(partSize))), nil
}

func StatusToString(status int32) string {
	switch status {
	case JobStatusNotRunning:
		return "NOT_RUNNING"
	case JobStatusStarting:
		return "STARTING"
	case JobStatusRunning:
		return "RUNNING"
	case JobStatusSuspended:
		return "SUSPENDED"
	case JobStatusSuspendedExportingSnapshot:
		return "SUSPENDED_EXPORTING_SNAPSHOT"
	case JobStatusCompleting:
		return "COMPLETING"
	case JobStatusFailed:
		return "FAILED"
	case JobStatusCompleted:
		return "COMPLETED"
	}
	return "UNKNOWN"
}

type binBatch struct {
	reader io.Reader
	buf    []byte
}

func newBatch(reader io.Reader, batchSize int) *binBatch {
	if batchSize < 1 {
		panic("newBatch: batchSize must be positive")
	}
	return &binBatch{
		reader: reader,
		buf:    make([]byte, batchSize),
	}
}

// Next returns the next batch of bytes.
// Make sure to copy it before calling Next again.
func (bb *binBatch) Next() ([]byte, []byte, error) {
	n, err := bb.reader.Read(bb.buf)
	if err != nil {
		return nil, nil, err
	}
	b := bb.buf[0:n:n]
	h, err := calculateHash(bytes.NewBuffer(b))
	if err != nil {
		return nil, nil, err
	}
	return b, h, nil
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

func MakeJobNameIDMaps(jobList []control.JobAndSqlSummary) (jobNameToID map[string]int64, idToJobName map[int64]string) {
	jobNameToID = make(map[string]int64, len(jobList))
	idToJobName = make(map[int64]string, len(jobList))
	for _, j := range jobList {
		idToJobName[j.JobId] = j.NameOrId
		if j.Status == JobStatusFailed || j.Status == JobStatusCompleted {
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
		return GetJobList(ctx, ci)
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

func GetJobList(ctx context.Context, ci *hazelcast.ClientInternal) ([]control.JobAndSqlSummary, error) {
	req := codec.EncodeJetGetJobAndSqlSummaryListRequest()
	resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	ls := codec.DecodeJetGetJobAndSqlSummaryListResponse(resp)
	return ls, nil
}
