package snapshot

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
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
	})
	if err != nil {
		return err
	}
	stop()
	r := rows.([]output.Row)
	if len(r) == 0 {
		ec.SetResultString("No snapshots found")
		return nil
	}
	return ec.AddOutputRows(ctx, r...)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("snapshot:list", ListCmd{}))
}
