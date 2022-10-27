package base

import (
	"context"
	"io"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type DelimitedPrinter struct{}

func (pr DelimitedPrinter) Print(w io.Writer, rows []output.Row) error {
	result := output.NewSimpleRows(rows)
	dr := output.NewDelimitedResult("\t", result, true)
	_, err := dr.Serialize(context.Background(), w)
	return err
}

type JSONPrinter struct{}

func (pr JSONPrinter) Print(w io.Writer, rows []output.Row) error {
	result := output.NewSimpleRows(rows)
	jr := output.NewJSONResult(result)
	_, err := jr.Serialize(context.Background(), w)
	return err
}

type TablePrinter struct {
	Mode          output.TableOutputMode
	headerPrinted bool
}

func (pr *TablePrinter) Print(w io.Writer, rows []output.Row) error {
	header, rows := output.MakeTable(rows)
	result := output.NewSimpleRows(rows)
	tr := output.NewTableResult(header, result)
	_, err := tr.Serialize(context.Background(), w, pr.Mode)
	return err
}

func init() {
	plug.Registry.RegisterPrinter("delimited", &DelimitedPrinter{})
	plug.Registry.RegisterPrinter("json", &JSONPrinter{})
	plug.Registry.RegisterPrinter("table", &TablePrinter{})
	plug.Registry.RegisterPrinter("csv", &TablePrinter{Mode: output.TableOutputModeCSV})
}
