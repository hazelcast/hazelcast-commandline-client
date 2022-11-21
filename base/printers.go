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

func (pr DelimitedPrinter) Print(w io.Writer, rp output.RowProducer) error {
	dr := output.NewDelimitedResult("\t", rp, true)
	_, err := dr.Serialize(context.Background(), w)
	return err
}

type JSONPrinter struct{}

func (pr JSONPrinter) Print(w io.Writer, rp output.RowProducer) error {
	jr := output.NewJSONResult(rp)
	_, err := jr.Serialize(context.Background(), w)
	return err
}

type TablePrinter struct {
	Mode output.TableOutputMode
}

func (pr *TablePrinter) Print(w io.Writer, rp output.RowProducer) error {
	rows := output.MaterializeRows(rp)
	header, rows := output.MakeTable(rows)
	rp = output.NewSimpleRows(rows)
	tr := output.NewTableResult(header, rp)
	_, err := tr.Serialize(context.Background(), w)
	return err
}

func init() {
	plug.Registry.RegisterPrinter(PrinterDelimited, &DelimitedPrinter{})
	plug.Registry.RegisterPrinter(PrinterJSON, &JSONPrinter{})
	plug.Registry.RegisterPrinter(PrinterTable, &TablePrinter{})
	plug.Registry.RegisterPrinter(PrinterCSV, &TablePrinter{Mode: output.TableOutputModeCSV})
}
