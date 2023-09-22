//go:build std || job

package job

import (
	"context"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/jet"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListCommand struct{}

func (ListCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list")
	help := "List jobs"
	cc.SetCommandHelp(help, help)
	cc.AddBoolFlag(flagIncludeSQL, "", false, false, "include SQL jobs")
	cc.AddBoolFlag(flagIncludeUserCancelled, "", false, false, "include user cancelled jobs")
	return nil
}

func (ListCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	rows, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) ([]output.Row, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metrics.NewKey(cid, vid), "total.job."+cmd.RunningMode(ec))
		sp.SetText("Getting the job list")
		j := jet.New(ci, sp, ec.Logger())
		jl, err := j.GetJobList(ctx)
		if err != nil {
			return nil, err
		}
		return createJetJobRows(ec, jl), nil
	})
	if err != nil {
		return err
	}
	stop()
	if len(rows) == 0 {
		ec.PrintlnUnnecessary("OK No jobs found.")
		return nil
	}
	return ec.AddOutputRows(ctx, rows...)
}

func createJetJobRows(ec plug.ExecContext, lsi interface{}) []output.Row {
	ls := lsi.([]control.JobAndSqlSummary)
	rows := make([]output.Row, 0, len(ls))
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	sql := ec.Props().GetBool(flagIncludeSQL)
	user := ec.Props().GetBool(flagIncludeUserCancelled)
	for _, v := range ls {
		if !sql && v.SqlSummary.Query != "" {
			// this is an SQL job but --include-sql was not used, so skip it
			continue
		}
		if !user && v.UserCancelled {
			// this is a user cancelled job but --include-user-cancelled was not used, so skip it
			continue
		}
		id := idToString(v.JobId)
		name := v.NameOrId
		if name == id {
			name = "N/A"
		}
		status := jet.StatusToString(v.Status)
		if status == "FAILED" && v.UserCancelled {
			status = "CANCELLED"
		}
		row := output.Row{
			output.Column{
				Name:  "Job ID",
				Type:  serialization.TypeString,
				Value: id,
			},
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: name,
			},
			output.Column{
				Name:  "Status",
				Type:  serialization.TypeString,
				Value: status,
			},
			msToOffsetDateTimeColumn(v.SubmissionTime, "Submitted"),
			msToOffsetDateTimeColumn(v.CompletionTime, "Completed"),
		}
		if sql {
			row = append(row, output.Column{
				Name:  "Query",
				Type:  serialization.TypeString,
				Value: v.SqlSummary.Query,
			}, output.Column{
				Name:  "Unbounded",
				Type:  serialization.TypeBool,
				Value: v.SqlSummary.Unbounded,
			})
		}
		if verbose {
			row = append(row, output.Column{
				Name:  "Light",
				Type:  serialization.TypeBool,
				Value: v.LightJob,
			}, output.Column{
				Name:  "Message",
				Type:  serialization.TypeString,
				Value: v.FailureText,
			})
		}
		rows = append(rows, row)
	}
	return rows
}

func msToOffsetDateTimeColumn(ms int64, name string) output.Column {
	if ms == 0 {
		return output.Column{
			Name: name,
			Type: serialization.TypeNil,
		}
	}
	return output.Column{
		Name:  name,
		Type:  serialization.TypeJavaLocalDateTime,
		Value: types.LocalDateTime(time.UnixMilli(ms)),
	}
}

func init() {
	check.Must(plug.Registry.RegisterCommand("job:list", &ListCommand{}))
}
