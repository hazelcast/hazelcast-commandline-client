package sql

import (
	"errors"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client/sql"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func updateOutput(ec plug.ExecContext, res sql.Result, verbose bool) error {
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
		if err := ec.FlushOutput(); err != nil {
			return err
		}
	}
	return nil
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
