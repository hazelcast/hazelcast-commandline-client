package job

import (
	"context"
	"crypto/sha256"
	"fmt"
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
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

const (
	// see: https://github.com/hazelcast/hazelcast/issues/24285
	envExperimentalCalculateHashWorkaround = "CLC_EXPERIMENTAL_WORKAROUND_24285"
	minServerVersion                       = "5.3.0-BETA-2"
)

type SubmitCmd struct{}

func (cm SubmitCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("submit [jar file] [arg, ...]")
	long := fmt.Sprintf(`Submit a jar file and create a Jet job
	
This command requires a Viridian or a Hazelcast cluster
having version %s or better.
`, minServerVersion)
	short := "Submit a jar file and create a Jet job"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(flagName, "", "", false, "override the job name")
	cc.AddStringFlag(flagSnapshot, "", "", false, "initial snapshot to start the job from")
	cc.AddStringFlag(flagClass, "", "", false, "set the main class")
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
	// TODO: split the binary
	sid := types.NewUUID()
	bin, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	_, fn := filepath.Split(path)
	fn = strings.TrimSuffix(fn, ".jar")
	args := ec.Args()[1:]
	jobName := ec.Props().GetString(flagName)
	snapshot := ec.Props().GetString(flagSnapshot)
	className := ec.Props().GetString(flagClass)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Uploading metadata")
		hash, workaround := calculateHashWithWorkaround(ci, bin)
		if workaround {
			ec.Logger().Info("Working around https://github.com/hazelcast/hazelcast/issues/24285")
		}
		req := codec.EncodeJetUploadJobMetaDataRequest(sid, false, fn, hash, snapshot, jobName, className, args)
		mem, err := randomMember(ctx, ci)
		if err != nil {
			return nil, fmt.Errorf("uploading job metadata: %w", err)
		}
		if _, err = ci.InvokeOnMember(ctx, req, mem, nil); err != nil {
			return nil, err
		}
		sp.SetText("Uploading Jar")
		req = codec.EncodeJetUploadJobMultipartRequest(sid, 1, 1, bin, int32(len(bin)), hash)
		if _, err = ci.InvokeOnMember(ctx, req, mem, nil); err != nil {
			return nil, fmt.Errorf("uploading jar file: %w", err)
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("uploading metadata: %w", err)
	}
	stop()
	return nil
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
		time.Sleep(100 * time.Millisecond)
	}
}

func calculateHash(bin []byte) string {
	// use the following when the member-side is fixed
	return fmt.Sprintf("%x", sha256.Sum256(bin))
}

func calculateHashWithWorkaround(ci *hazelcast.ClientInternal, bin []byte) (string, bool) {
	var workaround bool
	w := os.Getenv(envExperimentalCalculateHashWorkaround)
	if w == "1" {
		workaround = true
	}
	if cmd.ServerVersionOf(ci) == "5.3.0-BETA-2" {
		workaround = true
	}
	hash := calculateHash(bin)
	if workaround && hash[0] == '0' {
		hash = hash[1:]
	}
	return hash, workaround
}

func init() {
	Must(plug.Registry.RegisterCommand("job:submit", &SubmitCmd{}))
}
