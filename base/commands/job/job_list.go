package job

import (
	"context"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec/control"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListCmd struct{}

func (cm ListCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("list")
	help := "List jobs"
	cc.SetCommandHelp(help, help)
	cc.SetPositionalArgCount(0, 0)
	cc.AddBoolFlag(flagIncludeSQL, "", false, false, "include SQL jobs")
	cc.AddBoolFlag(flagIncludeUserCancelled, "", false, false, "include user cancelled jobs")
	return nil
}

func (cm ListCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	ls, cancel, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("Getting the job list")
		return getJobList(ctx, ci)
	})
	if err != nil {
		return err
	}
	defer cancel()
	return outputJetJobs(ctx, ec, ls)
}

func outputJetJobs(ctx context.Context, ec plug.ExecContext, lsi interface{}) error {
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
		row := output.Row{
			output.Column{
				Name:  "Job ID",
				Type:  serialization.TypeString,
				Value: idToString(v.JobId),
			},
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: v.NameOrId,
			},
			output.Column{
				Name:  "Status",
				Type:  serialization.TypeString,
				Value: statusToString(v.Status),
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
		if user {
			row = append(row, output.Column{
				Name:  "User Cancelled",
				Type:  serialization.TypeBool,
				Value: v.UserCancelled,
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
	if len(rows) == 0 {
		ec.PrintlnUnnecessary("There are no jobs.")
	}
	return ec.AddOutputRows(ctx, rows...)
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

func statusToString(status int32) string {
	switch status {
	case statusNotRunning:
		return "NOT_RUNNING"
	case statusStarting:
		return "STARTING"
	case statusRunning:
		return "RUNNING"
	case statusSuspended:
		return "SUSPENDED"
	case statusSuspendedExportingSnapshot:
		return "SUSPENDED_EXPORTING_SNAPSHOT"
	case statusCompleting:
		return "COMPLETING"
	case statusFailed:
		return "FAILED"
	case statusCompleted:
		return "COMPLETED"
	}
	return "UNKNOWN"
}

func init() {
	Must(plug.Registry.RegisterCommand("job:list", &ListCmd{}))
}
