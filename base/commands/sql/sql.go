package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	propertyApplySuggestion = "apply-suggestion"
)

type SQLCommand struct{}

func (cm *SQLCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("sql QUERY [flags]")
	cc.SetPositionalArgCount(0, 1)
	cc.AddCommandGroup("sql", "SQL")
	cc.SetCommandGroup("sql")
	long := `Run the given SQL query

If COMMAND is shell, then the SQL shell is started.
`
	cc.SetCommandHelp(long, "Run SQL")
	cc.AddBoolFlag(propertyApplySuggestion, "", false, false, "execute the proposed CREATE MAPPING suggestion and retry the query")
	return nil
}

func (cm *SQLCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	if len(ec.Args()) < 1 {
		ec.ShowHelpAndExit()
		return nil
	}
	query := ec.Args()[0]
	res, err := cm.execQuery(ctx, query, ec)
	if err != nil {
		return err
	}
	// TODO: keep it or remove it?
	verbose := ec.Props().GetBool(clc.PropertyVerbose)
	return updateOutput(ec, res, verbose)
}

func (cm *SQLCommand) execQuery(ctx context.Context, query string, ec plug.ExecContext) (sql.Result, error) {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, err
	}
	as := ec.Props().GetBool(propertyApplySuggestion)
	r, err := execSQL(ctx, ec, ci, query)
	if err != nil {
		// check whether this is an SQL error with a suggestion,
		// so we can improve the error message or apply the suggestion if there's one
		var serr *sql.Error
		if !errors.As(err, &serr) {
			return nil, err
		}
		// TODO: This changes the error in order to remove 'decoding SQL execute response:' prefix.
		// Once that is removed from the Go client, the code below may be removed.
		err = adaptSQLError(err)
		if !as {
			if serr.Suggestion != "" {
				return nil, fmt.Errorf("%w\n\nUse --%s to automatically apply the suggestion", err, propertyApplySuggestion)
			}
			return nil, err
		}
		if serr.Suggestion != "" {
			ec.Logger().Debug(func() string {
				return fmt.Sprintf("Re-trying executing SQL with suggestion: %s", serr.Suggestion)
			})
			// execute the suggested query
			if _, err := ci.Client().SQL().Execute(ctx, serr.Suggestion); err != nil {
				return nil, err
			}
			// execute the original query
			return ci.Client().SQL().Execute(ctx, query)
		}
	}
	return r, err
}

func convertSQLType(ct sql.ColumnType) int32 {
	switch ct {
	case sql.ColumnTypeVarchar:
		return serialization.TypeString
	case sql.ColumnTypeBoolean:
		return serialization.TypeBool
	case sql.ColumnTypeTinyInt:
		return serialization.TypeByte
	case sql.ColumnTypeSmallInt:
		return serialization.TypeInt16
	case sql.ColumnTypeInt:
		return serialization.TypeInt32
	case sql.ColumnTypeBigInt:
		return serialization.TypeInt64
	case sql.ColumnTypeDecimal:
		return serialization.TypeJavaDecimal
	case sql.ColumnTypeReal:
		return serialization.TypeFloat32
	case sql.ColumnTypeDouble:
		return serialization.TypeFloat64
	case sql.ColumnTypeDate:
		return serialization.TypeJavaLocalDate
	case sql.ColumnTypeTime:
		return serialization.TypeJavaLocalTime
	case sql.ColumnTypeTimestamp:
		return serialization.TypeJavaLocalDateTime
	case sql.ColumnTypeTimestampWithTimeZone:
		return serialization.TypeJavaOffsetDateTime
	case sql.ColumnTypeObject:
		return serialization.TypeSkip
	case sql.ColumnTypeNull:
		return serialization.TypeNil
	case sql.ColumnTypeJSON:
		return serialization.TypeJSONSerialization
	}
	return serialization.TypeNotDecoded
}

func init() {
	Must(plug.Registry.RegisterCommand("sql", &SQLCommand{}))
}
