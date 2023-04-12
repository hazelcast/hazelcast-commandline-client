package sql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
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

func UpdateOutput(ctx context.Context, ec plug.ExecContext, res sql.Result, verbose bool) error {
	if !res.IsRowSet() {
		return nil
	}
	it, err := res.Iterator()
	if err != nil {
		return err
	}
	rowCh := make(chan output.Row, 1)
	errCh := make(chan error)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	go func() {
		cols := MustValue(res.RowMetadata()).Columns()
		var row sql.Row
		var err error
		for it.HasNext() {
			row, err = it.Next()
			if err != nil {
				break
			}
			// have to create a new output row
			// since it is processed by another goroutine
			orow := make(output.Row, len(cols))
			for i, col := range cols {
				orow[i] = output.Column{
					Name:  col.Name(),
					Type:  convertSQLType(col.Type()),
					Value: MustValue(row.Get(i)),
				}
			}
			select {
			case rowCh <- orow:
			case <-ctx.Done():
				break
			}
		}
		close(rowCh)
		errCh <- err
	}()
	_ = ec.AddOutputStream(ctx, rowCh)
	select {
	case err = <-errCh:
		return err
	case <-ctx.Done():
		break
	}
	return nil
}

func ExecSQL(ctx context.Context, ec plug.ExecContext, query string) (sql.Result, context.CancelFunc, error) {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, nil, err
	}
	as := ec.Props().GetBool(propertyUseMappingSuggestion)
	rv, stop, err := execSQL(ctx, ec, ci, query)
	if err != nil {
		// check whether this is an SQL error with a suggestion,
		// so we can improve the error message or apply the suggestion if there's one
		var serr *sql.Error
		if !errors.As(err, &serr) {
			return nil, stop, err
		}
		// TODO: This changes the error in order to remove 'decoding SQL execute response:' prefix.
		// Once that is removed from the Go client, the code below may be removed.
		err = AdaptSQLError(err)
		if !as {
			if serr.Suggestion != "" && !ec.Interactive() {
				return nil, stop, fmt.Errorf("%w\n\nUse --%s to automatically apply the suggestion", err, propertyUseMappingSuggestion)
			}
			return nil, stop, err
		}
		if serr.Suggestion != "" {
			ec.Logger().Debug(func() string {
				return fmt.Sprintf("Re-trying executing SQL with suggestion: %s", serr.Suggestion)
			})
			// execute the suggested query
			_, stop, err := execSQL(ctx, ec, ci, serr.Suggestion)
			if err != nil {
				return nil, stop, err
			}
			stop()
			// execute the original query
			return execSQL(ctx, ec, ci, query)
		}
	}
	return rv, stop, nil
}

func execSQL(ctx context.Context, ec plug.ExecContext, ci *hazelcast.ClientInternal, query string) (sql.Result, context.CancelFunc, error) {
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
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
	})
	if err != nil {
		return nil, stop, err
	}
	return rv.(sql.Result), stop, nil
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
