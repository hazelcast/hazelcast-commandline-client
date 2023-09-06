//go:build std || snapshot

package snapshot

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	argSnapshotName      = "snapshotName"
	argTitleSnapshotName = "snapshot name"
)

type DeleteCmd struct{}

func (cm DeleteCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("delete")
	help := "Delete a snapshot"
	cc.SetCommandHelp(help, help)
	cc.AddStringArg(argSnapshotName, argTitleSnapshotName)
	return nil
}

func (cm DeleteCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	name := ec.GetStringArg(argTitleSnapshotName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Deleting the snapshot")
		sm, err := ci.Client().GetMap(ctx, jetExportedSnapshotsMap)
		if err != nil {
			return nil, err
		}
		if err := sm.Delete(ctx, name); err != nil {
			return nil, err
		}
		m, err := ci.Client().GetMap(ctx, jetExportedSnapshotPrefix+name)
		if err != nil {
			return nil, err
		}
		return nil, m.Destroy(ctx)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("snapshot:delete", DeleteCmd{}))
}
