//go:build base || sql

package sql

import (
	"context"

	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	propertyUseMappingSuggestion = "use-mapping-suggestion"
)

type SQLCommand struct{}

func (cm *SQLCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	// set the default format to table in the interactive mode
	if ec.CommandName() == "clc shell" && len(ec.Args()) == 0 {
		props.Set(clc.PropertyFormat, base.PrinterTable)
	}
	return nil
}

func (cm *SQLCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("sql [QUERY] [flags]")
	cc.SetPositionalArgCount(1, 1)
	cc.AddCommandGroup("sql", "SQL")
	cc.SetCommandGroup("sql")
	long := `Runs the given SQL query or starts the SQL shell

If QUERY is not given, then the SQL shell is started.
`
	cc.SetCommandHelp(long, "Run SQL")
	cc.AddBoolFlag(propertyUseMappingSuggestion, "", false, false, "execute the proposed CREATE MAPPING suggestion and retry the query")
	return nil
}

func (cm *SQLCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	// this method is only for the non-interactive mode
	if len(ec.Args()) < 1 {
		return nil
	}
	query := ec.Args()[0]
	res, stop, err := cm.execQuery(ctx, query, ec)
	defer stop()
	if err != nil {
		return err
	}
	// TODO: keep it or remove it?
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	return UpdateOutput(ctx, ec, res, verbose)
}

func (cm *SQLCommand) execQuery(ctx context.Context, query string, ec plug.ExecContext) (sql.Result, context.CancelFunc, error) {
	return ExecSQL(ctx, ec, query)
}

func init() {
	plug.Registry.RegisterAugmentor("20-sql", &SQLCommand{})
	Must(plug.Registry.RegisterCommand("sql", &SQLCommand{}))
}
