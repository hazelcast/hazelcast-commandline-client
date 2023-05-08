package job

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

const (
	// see: https://github.com/hazelcast/hazelcast/issues/24285
	envExperimentalCalculateHashWorkaround = "CLC_EXPERIMENTAL_WORKAROUND_24285"
	minServerVersion                       = "5.3.0-BETA-2"
	defaultBatchSize                       = 2 * 1024 * 1024 // 10MB
)

type SubmitCmd struct{}

func (cm SubmitCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("submit [jar-file] [arg, ...]")
	long := fmt.Sprintf(`Submits a jar file to create a Jet job
	
This command requires a Viridian or a Hazelcast cluster
having version %s or better.
`, minServerVersion)
	short := "Submits a jar file to create a Jet job"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(flagName, "", "", false, "override the job name")
	cc.AddStringFlag(flagSnapshot, "", "", false, "initial snapshot to start the job from")
	cc.AddStringFlag(flagClass, "", "", false, "the class that contains the main method that creates the Jet job")
	cc.AddIntFlag(flagRetries, "", 3, false, "number of times to retry a failed upload attempt")
	cc.AddBoolFlag(flagWait, "", false, false, "wait for the job to transition to RUNNING state")
	cc.SetPositionalArgCount(1, math.MaxInt)
	return nil
}

func (cm SubmitCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	path := ec.Args()[0]
	if !paths.Exists(path) {
		return fmt.Errorf("file does not exists: %s", path)
	}
	if !strings.HasSuffix(path, ".jar") {
		return fmt.Errorf("submitted file is not a jar file: %s", path)
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	if sv, ok := cmd.CheckServerCompatible(ci, minServerVersion); !ok {
		return fmt.Errorf("server (%s) does not support this command, at least %s is expected", sv, minServerVersion)
	}
	return submitJar(ctx, ci, ec, path)
}

func submitJar(ctx context.Context, ci *hazelcast.ClientInternal, ec plug.ExecContext, path string) error {
	wait := ec.Props().GetBool(flagWait)
	jobName := ec.Props().GetString(flagName)
	snapshot := ec.Props().GetString(flagSnapshot)
	className := ec.Props().GetString(flagClass)
	if wait && jobName == "" {
		return fmt.Errorf("--wait requires the --name to be set")
	}
	tries := int(ec.Props().GetInt(flagRetries))
	if tries < 0 {
		tries = 0
	}
	tries++
	sid := types.NewUUID()
	_, fn := filepath.Split(path)
	fn = strings.TrimSuffix(fn, ".jar")
	args := ec.Args()[1:]
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		workaround := workaround24285(ci)
		if workaround {
			ec.Logger().Debugf("Working around https://github.com/hazelcast/hazelcast/issues/24285")
		}
		hashBin, err := hashOfPath(path)
		if err != nil {
			return nil, err
		}
		hash := hashWithWorkaround(hashBin, workaround)
		mrReq := codec.EncodeJetUploadJobMetaDataRequest(sid, false, fn, hash, snapshot, jobName, className, args)
		err = retry(tries, ec.Logger(), func() error {
			sp.SetProgress(0)
			mem, err := randomMember(ctx, ci)
			if err != nil {
				return err
			}
			sp.SetText("Uploading the metadata")
			if _, err = ci.InvokeOnMember(ctx, mrReq, mem, nil); err != nil {
				return err
			}
			if err != nil {
				return fmt.Errorf("uploading job metadata: %w", err)
			}
			sp.SetText("Uploading the job")
			pc, err := partCountOf(path, defaultBatchSize)
			if err != nil {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			ec.Logger().Info("Sending %s in %d batch(es)", path, pc)
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
				hash := hashWithWorkaround(hashBin, workaround)
				mrReq = codec.EncodeJetUploadJobMultipartRequest(sid, part, int32(pc), bin, int32(len(bin)), hash)
				if _, err := ci.InvokeOnMember(ctx, mrReq, mem, nil); err != nil {
					return fmt.Errorf("uploading part %d: %w", part, err)
				}
				sp.SetProgress(float32(part) / float32(pc))
			}
			sp.SetProgress(1)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("uploading the job: %w", err)
	}
	stop()
	if wait {
		msg := fmt.Sprintf("Waiting for job %s to transition to RUNNING state", jobName)
		ec.Logger().Info(msg)
		err = WaitJobState(ctx, ec, msg, jobName, statusRunning, 3*time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}

func partCountOf(path string, partSize int) (int, error) {
	st, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return int(math.Ceil(float64(st.Size()) / float64(partSize))), nil
}

func retry(times int, lg log.Logger, f func() error) error {
	var err error
	for i := 0; i < times; i++ {
		err = f()
		if err != nil {
			lg.Error(err)
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed after %d tries: %w", times, err)
}

func hashOfPath(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return calculateHash(f)
}

func randomMember(ctx context.Context, ci *hazelcast.ClientInternal) (types.UUID, error) {
	var mi cluster.MemberInfo
	for {
		if ctx.Err() != nil {
			return types.UUID{}, ctx.Err()
		}
		mems := ci.OrderedMembers()
		if len(mems) != 0 {
			mi = mems[rand.Intn(len(mems))]
			if ci.ConnectedToMember(mi.UUID) {
				return mi.UUID, nil
			}
		}
	}
}

func workaround24285(ci *hazelcast.ClientInternal) bool {
	if os.Getenv(envExperimentalCalculateHashWorkaround) == "1" {
		return true
	}
	return cmd.ServerVersionOf(ci) == "5.3.0-BETA-2"
}

func hashWithWorkaround(hash []byte, workaround bool) string {
	text := fmt.Sprintf("%x", hash)
	if workaround && text[0] == '0' {
		text = text[1:]
	}
	return text
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

func init() {
	Must(plug.Registry.RegisterCommand("job:submit", &SubmitCmd{}))
}
