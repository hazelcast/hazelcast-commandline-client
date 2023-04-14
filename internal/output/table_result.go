package output

import (
	"context"
	"io"

	"github.com/fatih/color"

	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

type TableResult struct {
	header   []table.Column
	rp       RowProducer
	maxWidth int
}

// NewTableResult creates a table result from the row producer and (optional) header.
// If header is not given, then it is assumed the first row in the row producer is the header, and alignment is auto-calculated.
// In this case maxWidth is required.
func NewTableResult(header []table.Column, rp RowProducer, maxWidth int) *TableResult {
	if header == nil && maxWidth <= 0 {
		panic("maxWidth should be positive if header is nil")
	}
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
			if tr.header != nil {
				t.WriteHeader(tr.header)
			} else {
				t.WriteHeader(makeTableHeaderFromRow(vr, tr.maxWidth))
			}
			wroteHeader = true
		}
		row := make([]string, len(vr))
		for i, v := range vr {
			row[i] = v.Text()
		}
		t.WriteRow(row)
	}
	t.End()
	return n, nil
}
