package job

import (
	"context"
	"errors"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListCmd struct{}

func (cm ListCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list")
	help := "List jobs"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm ListCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeJetGetJobSummaryListRequest()
	ls, cancel, err := ec.ExecuteBlocking(ctx, "Getting the job list", func(ctx context.Context) (any, error) {
		resp, err := ci.InvokeOnRandomTarget(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		data := codec.DecodeJetGetJobSummaryListResponse(resp)
		return ci.DecodeData(data)
	})
	if err != nil {
		return err
	}
	defer cancel()
	return outputJetJobs(ctx, ec, ls)
}

func outputJetJobs(ctx context.Context, ec plug.ExecContext, mi interface{}) error {
	m, ok := mi.([]interface{})
	if !ok {
		return errors.New("invalid JetGetJobIds response")
	}
	rows := make([]output.Row, 0, len(m))
	for _, vv := range m {
		v := vv.(*config.JetJobSummary)
		rows = append(rows, output.Row{
			output.Column{
				Name:  "ID",
				Type:  serialization.TypeInt64,
				Value: v.ID,
			},
			//output.Column{
			//	Name:  "Job ID",
			//	Type:  serialization.TypeInt64,
			//	Value: v.IDs[i],
			//},
			//output.Column{
			//	Name:  "Is Light",
			//	Type:  serialization.TypeBool,
			//	Value: v.Light[i],
			//},
		})
	}
	return ec.AddOutputRows(ctx, rows...)
}

func init() {
	Must(plug.Registry.RegisterCommand("job:list", &ListCmd{}))
}
