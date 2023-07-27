//go:build std || job

package job

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ExportSnapshotCmd struct{}

func (cm ExportSnapshotCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("export-snapshot [job-ID/name]")
	long := "Exports a snapshot for a job.\nThis feature requires a Viridian or Hazelcast Enterprise cluster."
	short := "Exports a snapshot for a job"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, 1)
	cc.AddStringFlag(flagName, "", "", false, "specify the snapshot. By default an auto-genertaed snapshot name is used")
	cc.AddBoolFlag(flagCancel, "", false, false, "cancel the job after taking the snapshot")
	return nil
}

func (cm ExportSnapshotCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	var jm *JobsInfo
	var jid int64
	var ok bool
	jobNameOrID := ec.Args()[0]
	name := ec.Props().GetString(flagName)
	cancel := ec.Props().GetBool(flagCancel)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
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
		sp.SetText(fmt.Sprintf("Exporting snapshot: %s", name))
		return nil, j.ExportSnapshot(ctx, jid, name, cancel)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func autoGenerateSnapshotName(jobName string) string {
	dt := time.Now().UTC().Format("06-01-02_150405")
	r := rand.Int31n(10_000)
	return fmt.Sprintf("%s-%s-%d", jobName, dt, r)
}

func init() {
	Must(plug.Registry.RegisterCommand("job:export-snapshot", &ExportSnapshotCmd{}))
}
