package sql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client/hzerrors"
	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func UpdateOutput(ec plug.ExecContext, res sql.Result, verbose bool) error {
	// we enable streaming only for non-table output
	// TODO: properly fix the table output
	f := ec.Props().GetString(clc.PropertyFormat)
	tableOutput := f != base.PrinterJSON && f != base.PrinterDelimited
	if !res.IsRowSet() {
		if verbose {
			ec.AddOutputRows(output.Row{
				{
					Name: "affected rows", Type: serialization.TypeInt64, Value: res.UpdateCount(),
				},
			})
		}
		return nil
	}
	it, err := res.Iterator()
	if err != nil {
		return err
	}
	for it.HasNext() {
		row, err := it.Next()
		if err != nil {
			return err
		}
		cols := row.Metadata().Columns()
		orow := make([]output.Column, len(cols))
		for i, col := range cols {
			orow[i] = output.Column{
				Name:  col.Name(),
				Type:  convertSQLType(col.Type()),
				Value: MustValue(row.Get(i)),
			}
		}
		ec.AddOutputRows(orow)
		if !tableOutput {
			if err := ec.FlushOutput(); err != nil {
				return err
			}
		}
	}
	if err := ec.FlushOutput(); err != nil {
		return err
	}
	return nil
}

func ExecSQL(ctx context.Context, ec plug.ExecContext, query string) (sql.Result, error) {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, err
	}
	as := ec.Props().GetBool(propertyUseMappingSuggestion)
	rv, err := ec.ExecuteBlocking(ctx, "Executing SQL", func(ctx context.Context) (any, error) {
		for {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			r, err := ci.Client().SQL().Execute(ctx, query)
			// If Go client cannot find a connection, it returns immediately with ErrIO
			// Retry logic here
			if err != nil {
				if errors.Is(err, hzerrors.ErrIO) {
					time.Sleep(1 * time.Second)
					continue
				}
				return nil, err
			}
			return r, nil
		}
	})
	if err != nil {
		// check whether this is an SQL error with a suggestion,
		// so we can improve the error message or apply the suggestion if there's one
		var serr *sql.Error
		if !errors.As(err, &serr) {
			return nil, err
		}
		// TODO: This changes the error in order to remove 'decoding SQL execute response:' prefix.
		// Once that is removed from the Go client, the code below may be removed.
		err = AdaptSQLError(err)
		if !as {
			if serr.Suggestion != "" {
				return nil, fmt.Errorf("%w\n\nUse --%s to automatically apply the suggestion", err, propertyUseMappingSuggestion)
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
	return rv.(sql.Result), nil
}

func AdaptSQLError(err error) error {
	var serr *sql.Error
	if !errors.As(err, &serr) {
		return err
	}
	// TODO: This changes the error in order to remove 'decoding SQL execute response:' prefix.
	// Once that is removed from the Go client, the code below may be removed.
	return fmt.Errorf(serr.Message)
}

var sqlTypeToSerializationType = map[sql.ColumnType]int32{
	sql.ColumnTypeVarchar:               serialization.TypeString,
	sql.ColumnTypeBoolean:               serialization.TypeBool,
	sql.ColumnTypeTinyInt:               serialization.TypeByte,
	sql.ColumnTypeSmallInt:              serialization.TypeInt16,
	sql.ColumnTypeInt:                   serialization.TypeInt32,
	sql.ColumnTypeBigInt:                serialization.TypeInt64,
	sql.ColumnTypeDecimal:               serialization.TypeJavaDecimal,
	sql.ColumnTypeReal:                  serialization.TypeFloat32,
	sql.ColumnTypeDouble:                serialization.TypeFloat64,
	sql.ColumnTypeDate:                  serialization.TypeJavaLocalDate,
	sql.ColumnTypeTime:                  serialization.TypeJavaLocalTime,
	sql.ColumnTypeTimestamp:             serialization.TypeJavaLocalDateTime,
	sql.ColumnTypeTimestampWithTimeZone: serialization.TypeJavaOffsetDateTime,
	sql.ColumnTypeObject:                serialization.TypeSkip,
	sql.ColumnTypeNull:                  serialization.TypeNil,
	sql.ColumnTypeJSON:                  serialization.TypeJSONSerialization,
}

func convertSQLType(ct sql.ColumnType) int32 {
	t, ok := sqlTypeToSerializationType[ct]
	if !ok {
		return serialization.TypeNotDecoded
	}
	return t
}
