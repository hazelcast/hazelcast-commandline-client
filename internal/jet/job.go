package jet

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/internal/cluster"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
)

type spinner interface {
	SetProgress(progress float32)
}

type Jet struct {
	ci *hazelcast.ClientInternal
	sp spinner
	lg log.Logger
}

func New(ci *hazelcast.ClientInternal, sp spinner, lg log.Logger) *Jet {
	return &Jet{
		ci: ci,
		sp: sp,
		lg: lg,
	}
}

func (j Jet) SubmitJob(ctx context.Context, path, jobName, className, snapshot string, args []string) error {
	_, fn := filepath.Split(path)
	fn = strings.TrimSuffix(fn, ".jar")
	j.sp.SetProgress(0)
	hashBin, err := hashOfPath(path)
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", hashBin)
	sid := types.NewUUID()
	mrReq := codec.EncodeJetUploadJobMetaDataRequest(sid, false, fn, hash, snapshot, jobName, className, args)
	mem, err := cluster.RandomMember(ctx, j.ci)
	if err != nil {
		return err
	}
	if _, err = j.ci.InvokeOnMember(ctx, mrReq, mem, nil); err != nil {
		return err
	}
	if err != nil {
		return fmt.Errorf("uploading job metadata: %w", err)
	}
	pc, err := partCountOf(path, defaultBatchSize)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	j.lg.Info("Sending %s in %d batch(es)", path, pc)
	bb := newBatch(f, defaultBatchSize)
	for i := int32(0); i < int32(pc); i++ {
		bin, hashBin, err := bb.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("sending the job: %w", err)
		}
		part := i + 1
		hash := fmt.Sprintf("%x", hashBin)
		mrReq = codec.EncodeJetUploadJobMultipartRequest(sid, part, int32(pc), bin, int32(len(bin)), hash)
		if _, err := j.ci.InvokeOnMember(ctx, mrReq, mem, nil); err != nil {
			return fmt.Errorf("uploading part %d: %w", part, err)
		}
		j.sp.SetProgress(float32(part) / float32(pc))
	}
	j.sp.SetProgress(1)
	return nil
}

func EnsureJobState(jobs []control.JobAndSqlSummary, jobNameOrID string, state int32) (bool, error) {
	for _, j := range jobs {
		if j.NameOrId == jobNameOrID {
			if j.Status == state {
				return true, nil
			}
			if j.Status == JobStatusFailed {
				return false, ErrJobFailed
			}
			return false, nil
		}
	}
	return false, ErrJobNotFound
}
