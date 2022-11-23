package output

import (
	"context"
	"fmt"
	"io"

	"github.com/fatih/color"

	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

type TableResult struct {
	header   []table.Column
	rp       RowProducer
	maxWidth int
}

func NewTableResult(header []table.Column, rp RowProducer, maxWidth int) *TableResult {
	return &TableResult{
		header:   header,
		rp:       rp,
		maxWidth: maxWidth,
	}
}

func (tr *TableResult) Serialize(ctx context.Context, w io.Writer) (int, error) {
	var n int
	cfg := table.Config{
		Stdout:     w,
		CellFormat: [2]string{" %s ", "| %s "},
	}
	// use the header separator if color is not enabled
	if color.NoColor {
		cfg.HeaderSeperator = "-"
	}
	t := table.New(cfg)
	wroteHeader := false
	for {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		vr, ok := tr.rp.NextRow(ctx)
		if !ok {
			break
		}
		if !wroteHeader {
			t.WriteHeader(makeTableHeaderFromRow(vr, tr.maxWidth))
			wroteHeader = true
		}
		row := make([]string, len(vr))
		for i, v := range vr {
			row[i] = fmt.Sprint(convertColumn(v))
		}
		t.WriteRow(row)
	}
	return n, nil
}
