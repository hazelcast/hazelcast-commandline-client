package base

import (
	"context"
	"io"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	PrinterDelimited = "delimited"
	PrinterJSON      = "json"
	PrinterTable     = "table"
	PrinterCSV       = "csv"
)

type DelimitedPrinter struct{}

func (pr DelimitedPrinter) PrintStream(ctx context.Context, w io.Writer, rp output.RowProducer) error {
	dr := output.NewDelimitedResult("\t", rp, true)
	_, err := dr.Serialize(ctx, w)
	return err
}

func (pr DelimitedPrinter) PrintRows(ctx context.Context, w io.Writer, rows []output.Row) error {
	rp := output.NewSimpleRows(rows)
	dr := output.NewDelimitedResult("\t", rp, true)
	_, err := dr.Serialize(ctx, w)
	return err
}

type JSONPrinter struct{}

func (pr JSONPrinter) PrintStream(ctx context.Context, w io.Writer, rp output.RowProducer) error {
	jr := output.NewJSONResult(rp)
	_, err := jr.Serialize(ctx, w)
	return err
}

func (pr JSONPrinter) PrintRows(ctx context.Context, w io.Writer, rows []output.Row) error {
	rp := output.NewSimpleRows(rows)
	jr := output.NewJSONResult(rp)
	_, err := jr.Serialize(ctx, w)
	return err
}

type TablePrinter struct {
	Mode output.TableOutputMode
}

func (pr *TablePrinter) PrintStream(ctx context.Context, w io.Writer, rp output.RowProducer) error {
	tr := output.NewTableResult(nil, rp)
	_, err := tr.Serialize(ctx, w)
	return err
}

func (pr *TablePrinter) PrintRows(ctx context.Context, w io.Writer, rows []output.Row) error {
	header, rows := output.MakeTableFromRows(rows)
	rp := output.NewSimpleRows(rows)
	tr := output.NewTableResult(header, rp)
	_, err := tr.Serialize(ctx, w)
	return err
}

func init() {
	plug.Registry.RegisterPrinter(PrinterDelimited, &DelimitedPrinter{})
	plug.Registry.RegisterPrinter(PrinterJSON, &JSONPrinter{})
	plug.Registry.RegisterPrinter(PrinterTable, &TablePrinter{})
	plug.Registry.RegisterPrinter(PrinterCSV, &TablePrinter{Mode: output.TableOutputModeCSV})
}
