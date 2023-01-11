package job

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type SubmitCmd struct{}

func (cm SubmitCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("submit [jar file]")
	help := "Submit a jar file and create a Jet job"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(1, 1)
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
	return submitJar(ctx, ci, ec, path)
}

func submitJar(ctx context.Context, ci *hazelcast.ClientInternal, ec plug.ExecContext, path string) error {
	// TODO: split the binary
	sid := types.NewUUID()
	bin, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(bin))
	_, fn := filepath.Split(path)
	fn = strings.TrimSuffix(fn, ".jar")
	req := codec.EncodeJetUploadJobMetaDataRequest(sid, fn, hash, "", "", "", nil)
	mi, cancel, err := ec.ExecuteBlocking(ctx, "Uploading metadata", func(ctx context.Context) (any, error) {
		mem, err := randomMember(ctx, ci)
		if err != nil {
			return nil, err
		}
		resp, err := ci.InvokeOnMember(ctx, req, mem, nil)
		if err != nil {
			return nil, err
		}
		ok := codec.DecodeJetUploadJobMetaDataResponse(resp)
		if !ok {
			return nil, errors.New("cannot upload job metadata")
		}
		return mem, nil
	})
	if err != nil {
		return fmt.Errorf("uploading metadata: %w", err)
	}
	defer cancel()
	mem := mi.(types.UUID)
	_, cancel, err = ec.ExecuteBlocking(ctx, "Uploading Jar", func(ctx context.Context) (any, error) {
		req = codec.EncodeJetUploadJobMultipartRequest(sid, 1, 1, bin, int32(len(bin)))
		resp, err := ci.InvokeOnMember(ctx, req, mem, nil)
		if err != nil {
			return nil, err
		}
		ok := codec.DecodeJetUploadJobMultipartResponse(resp)
		if !ok {
			return nil, errors.New("cannot upload the jar file")
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("uploading the jar file: %w", err)
	}
	defer cancel()
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

func init() {
	Must(plug.Registry.RegisterCommand("job:submit", &SubmitCmd{}))
}
