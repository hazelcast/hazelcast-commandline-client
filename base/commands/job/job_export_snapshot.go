//go:build std || job

package job

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	metric "github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ExportSnapshotCommand struct{}

func (ExportSnapshotCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("export-snapshot")
	long := "Exports a snapshot for a job.\nThis feature requires a Viridian or Hazelcast Enterprise cluster."
	short := "Exports a snapshot for a job"
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(flagName, "", "", false, "specify the snapshot. By default an auto-genertaed snapshot name is used")
	cc.AddBoolFlag(flagCancel, "", false, false, "cancel the job after taking the snapshot")
	cc.AddStringArg(argJobID, argTitleJobID)
	return nil
}

func (ExportSnapshotCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	var jm *JobsInfo
	var jid int64
	var ok bool
	jobNameOrID := ec.GetStringArg(argJobID)
	name := ec.Props().GetString(flagName)
	cancel := ec.Props().GetBool(flagCancel)
	row, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (output.Row, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metric.NewKey(cid, vid), "total.job."+cmd.RunningMode(ec))
		j := jet.New(ci, sp, ec.Logger())
		jis, err := j.GetJobList(ctx)
		if err != nil {
			return nil, err
		}
		if name == "" {
			// create the default snapshot name
			jm, err = NewJobNameToIDMap(jis)
			if err != nil {
				return nil, err
			}
			jid, ok = jm.GetIDForName(jobNameOrID)
			if !ok {
				return nil, jet.ErrInvalidJobID
			}
			info, ok := jm.GetInfoForID(jid)
			if !ok {
				name = "UNKNOWN"
			} else {
				name = autoGenerateSnapshotName(info.NameOrId)
			}
		} else {
			jm, err = NewJobNameToIDMap(jis)
			if err != nil {
				return nil, err
			}
			jid, ok = jm.GetIDForName(jobNameOrID)
			if !ok {
				return nil, jet.ErrInvalidJobID
			}
		}
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Exporting snapshot '%s'", name))
		if err := j.ExportSnapshot(ctx, jid, name, cancel); err != nil {
			return nil, err
		}
		row := output.Row{
			{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: name,
			},
		}
		return row, nil
	})
	if err != nil {
		return err
	}
	stop()
	ec.PrintlnUnnecessary("OK Exported the snapshot.\n")
	return ec.AddOutputRows(ctx, row)
}

func autoGenerateSnapshotName(jobName string) string {
	dt := time.Now().UTC().Format("06-01-02_150405")
	r := rand.Int31n(10_000)
	return fmt.Sprintf("%s-%s-%d", jobName, dt, r)
}

func init() {
	check.Must(plug.Registry.RegisterCommand("job:export-snapshot", &ExportSnapshotCommand{}))
}
