//go:build std || snapshot

package snapshot

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	argSnapshotName      = "snapshotName"
	argTitleSnapshotName = "snapshot name"
)

type DeleteCmd struct{}

func (DeleteCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("delete")
	help := "Delete a snapshot"
	cc.SetCommandHelp(help, help)
	cc.AddStringArg(argSnapshotName, argTitleSnapshotName)
	return nil
}

func (DeleteCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.GetStringArg(argTitleSnapshotName)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Deleting the snapshot '%s'", name))
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
	msg := fmt.Sprintf("OK Destroyed snapshot '%s'.", name)
	ec.PrintlnUnnecessary(msg)
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("snapshot:delete", DeleteCmd{}))
}
