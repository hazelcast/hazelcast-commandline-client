package base

import (
	"context"
	"io"
	"os"
	"strconv"

	"github.com/nathan-fiscaletti/consolesize-go"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
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

type TablePrinter struct{}

func (pr *TablePrinter) PrintStream(ctx context.Context, w io.Writer, rp output.RowProducer) error {
	mc, _ := consolesize.GetConsoleSize()
	if mc <= 0 {
		mc = maxCols()
	}
	tr := output.NewTableResult(nil, rp, mc)
	_, err := tr.Serialize(ctx, w)
	return err
}

func (pr *TablePrinter) PrintRows(ctx context.Context, w io.Writer, rows []output.Row) error {
	header, rows := output.MakeTableFromRows(rows)
	rp := output.NewSimpleRows(rows)
	tr := output.NewTableResult(header, rp, 0)
	_, err := tr.Serialize(ctx, w)
	return err
}

type CSVPrinter struct{}

func (pr *CSVPrinter) PrintStream(ctx context.Context, w io.Writer, rp output.RowProducer) error {
	cr := output.NewCSVResult(rp)
	_, err := cr.Serialize(ctx, w)
	return err
}

func (pr *CSVPrinter) PrintRows(ctx context.Context, w io.Writer, rows []output.Row) error {
	rp := output.NewSimpleRows(rows)
	cr := output.NewCSVResult(rp)
	_, err := cr.Serialize(ctx, w)
	return err
}

func maxCols() int {
	if s, ok := os.LookupEnv(clc.EnvMaxCols); ok {
		v, err := strconv.Atoi(s)
		if err == nil {
			return v
		}
	}
	return 1_000
}

func init() {
	plug.Registry.RegisterPrinter(PrinterDelimited, &DelimitedPrinter{})
	plug.Registry.RegisterPrinter(PrinterJSON, &JSONPrinter{})
	plug.Registry.RegisterPrinter(PrinterTable, &TablePrinter{})
	plug.Registry.RegisterPrinter(PrinterCSV, &CSVPrinter{})
}
