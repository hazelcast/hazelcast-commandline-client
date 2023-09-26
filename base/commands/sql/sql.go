//go:build std || sql

package sql

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	clcsql "github.com/hazelcast/hazelcast-commandline-client/clc/sql"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	minServerVersion = "5.0.0"
	argQuery         = "query"
	argTitleQuery    = "query"
)

type SQLCommand struct{}

func (SQLCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	// set the default format to table in the interactive mode
	if ecc, ok := ec.(clc.Arg0er); ok {
		if ec.CommandName() == ecc.Arg0()+" shell" && len(ec.Args()) == 0 {
			props.Set(clc.PropertyFormat, base.PrinterTable)
		}
	}
	return nil
}

func (SQLCommand) Init(cc plug.InitContext) error {
	if cc.Interactive() {
		return errors.ErrNotAvailable
	}
	cc.SetCommandUsage("sql")
	cc.AddCommandGroup("sql", "SQL")
	cc.SetCommandGroup("sql")
	long := fmt.Sprintf(`Runs the given SQL query or starts the SQL shell

If QUERY is not given, then the SQL shell is started.
	
This command requires a Viridian or a Hazelcast cluster
having version %s or better.
`, minServerVersion)
	cc.SetCommandHelp(long, "Run SQL")
	cc.AddBoolFlag(clcsql.PropertyUseMappingSuggestion, "", false, false, "execute the proposed CREATE MAPPING suggestion and retry the query")
	cc.AddStringArg(argQuery, argTitleQuery)
	return nil
}

func (SQLCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	// this method is only for the non-interactive mode
	if len(ec.Args()) < 1 {
		return nil
	}
	query := ec.GetStringArg(argQuery)
	resV, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		cmd.IncrementClusterMetric(ctx, ec, "total.sql")
		if sv, ok := cmd.CheckServerCompatible(ci, minServerVersion); !ok {
			return nil, fmt.Errorf("server (%s) does not support this command, at least %s is expected", sv, minServerVersion)
		}
		sp.SetText("Executing SQL")
		return clcsql.ExecSQL(ctx, ec, query)
	})
	if err != nil {
		return err
	}
	// this should be deferred because UpdateOutput will iterate on the result
	defer stop()
	res := resV.(sql.Result)
	return clcsql.UpdateOutput(ctx, ec, res)
}

func init() {
	plug.Registry.RegisterAugmentor("20-sql", &SQLCommand{})
	check.Must(plug.Registry.RegisterCommand("sql", &SQLCommand{}))
}
