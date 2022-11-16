package sql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/hzerrors"
	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func updateOutput(ec plug.ExecContext, res sql.Result, verbose bool) error {
	// we enable streaming only for non-table output
	// TODO: properly fix the table output
	f := ec.Props().GetString(clc.PropertyFormat)
	tableOutput := f != "json" && f != "delimited"
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

func execSQL(ctx context.Context, ec plug.ExecContext, ci *hazelcast.ClientInternal, query string) (sql.Result, error) {
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
		return nil, err
	}
	return rv.(sql.Result), nil
}

func adaptSQLError(err error) error {
	var serr *sql.Error
	if !errors.As(err, &serr) {
		return err
	}
	// TODO: This changes the error in order to remove 'decoding SQL execute response:' prefix.
	// Once that is removed from the Go client, the code below may be removed.
	return fmt.Errorf(serr.Message)
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
