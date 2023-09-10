package sql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/hazelcast/hazelcast-go-client/hzerrors"
	"github.com/hazelcast/hazelcast-go-client/sql"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const PropertyUseMappingSuggestion = "use-mapping-suggestion"

func ExecSQL(ctx context.Context, ec plug.ExecContext, sp clc.Spinner, query string) (sql.Result, error) {
	as := ec.Props().GetBool(PropertyUseMappingSuggestion)
	result, err := execSQL(ctx, ec, sp, query)
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
			if serr.Suggestion != "" && !ec.Interactive() {
				return nil, fmt.Errorf("%w\n\nUse --%s to automatically apply the suggestion", err, PropertyUseMappingSuggestion)
			}
			return nil, err
		}
		if serr.Suggestion != "" {
			ec.Logger().Debug(func() string {
				return fmt.Sprintf("Re-trying executing SQL with suggestion: %s", serr.Suggestion)
			})
			// execute the suggested query
			_, err := execSQL(ctx, ec, sp, serr.Suggestion)
			if err != nil {
				return nil, err
			}
			// execute the original query
			return execSQL(ctx, ec, sp, query)
		}
	}
	return result, nil
}

func execSQL(ctx context.Context, ec plug.ExecContext, sp clc.Spinner, query string) (sql.Result, error) {
	ci, err := cmd.ClientInternal(ctx, ec, sp)
	if err != nil {
		return nil, err
	}
	sp.SetText("Executing SQL")
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

func UpdateOutput(ctx context.Context, ec plug.ExecContext, res sql.Result) error {
	if !res.IsRowSet() {
		ec.PrintlnUnnecessary("OK Executed the query.")
		return nil
	}
	it, err := res.Iterator()
	if err != nil {
		return err
	}
	rowCh := make(chan output.Row, 1)
	errCh := make(chan error, 1)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	var count int64
	go func(count *int64) {
		var row sql.Row
		var err error
	loop:
		for it.HasNext() {
			row, err = it.Next()
			if err != nil {
				break
			}
			atomic.AddInt64(count, 1)
			// have to create a new output row
			// since it is processed by another goroutine
			cols := row.Metadata().Columns()
			orow := make(output.Row, len(cols))
			for i, col := range cols {
				orow[i] = output.Column{
					Name:  col.Name(),
					Type:  convertSQLType(col.Type()),
					Value: check.MustValue(row.Get(i)),
				}
			}
			select {
			case rowCh <- orow:
			case <-ctx.Done():
				break loop
			}
		}
		close(rowCh)
		errCh <- err
	}(&count)
	// XXX: the error is ignored, the reason must be noted.
	_ = ec.AddOutputStream(ctx, rowCh)
	select {
	case err = <-errCh:
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("OK Returned %d rows.", atomic.LoadInt64(&count))
		ec.PrintlnUnnecessary(msg)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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
