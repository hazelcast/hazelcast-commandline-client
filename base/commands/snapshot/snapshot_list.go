package snapshot

import (
	"context"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type ListCmd struct{}

func (cm ListCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list")
	help := "List snapshots"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (cm ListCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	rows, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Getting the snapshot list")
		m, err := ci.Client().GetMap(ctx, jetExportedSnapshotsMap)
		if err != nil {
			return nil, err
		}
		rows, err := listDetailRows(ctx, *m)
		if err != nil {
			ec.Logger().Error(err)
			rows, err = listRows(ctx, *m)
			if err != nil {
				return nil, err
			}
		}
		return rows, nil
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, rows.([]output.Row)...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("snapshot:list", ListCmd{}))
}

func listDetailRows(ctx context.Context, m hazelcast.Map) ([]output.Row, error) {
	esd, err := m.GetEntrySet(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]output.Row, 0, len(esd))
	for _, e := range esd {
		r := output.Row{}
		if s, ok := e.Key.(string); ok {
			r = append(r, output.Column{
				Name:  "Snapshot Name",
				Type:  serialization.TypeString,
				Value: s,
			})
		}
		if sd, ok := e.Value.(*serialization.Snapshot); ok {
			r = append(r, output.Column{
				Name:  "Job Name",
				Type:  serialization.TypeString,
				Value: sd.JobName,
			})
			r = append(r, output.Column{
				Name:  "Time",
				Type:  serialization.TypeJavaLocalDateTime,
				Value: types.LocalDateTime(time.UnixMilli(sd.CreationTime)),
			})
		}
		rows = append(rows, r)
	}
	return rows, nil
}

func listRows(ctx context.Context, m hazelcast.Map) ([]output.Row, error) {
	es, err := m.GetKeySet(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]output.Row, 0, len(es))
	for _, e := range es {
		if s, ok := e.(string); ok {
			rows = append(rows, output.Row{
				output.Column{
					Name:  "Snapshot Name",
					Type:  serialization.TypeString,
					Value: s,
				},
			})
		}
	}
	return rows, nil
}
