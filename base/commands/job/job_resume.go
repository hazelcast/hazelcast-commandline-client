package job

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

type ResumeCmd struct{}

func (cm ResumeCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("resume [job-ID/name]")
	help := "Resumes a suspended job"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (cm ResumeCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	jm, err := newJobNameToIDMap(ctx, ec, false)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		jid, ok := jm.GetIDForName(ec.Args()[0])
		if !ok {
			return nil, errInvalidJobID
		}
		sp.SetText(fmt.Sprintf("Resuming job: %s", idToString(jid)))
		req := codec.EncodeJetResumeJobRequest(jid)
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

func init() {
	Must(plug.Registry.RegisterCommand("job:resume", &ResumeCmd{}))
}
