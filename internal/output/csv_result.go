package output

import (
	"context"
	"encoding/csv"
	"io"
	"strings"
)

type CSVResult struct {
	rp RowProducer
}

func NewCSVResult(rp RowProducer) *CSVResult {
	return &CSVResult{
		rp: rp,
	}
}

func (cr *CSVResult) Serialize(ctx context.Context, w io.Writer) (int, error) {
	var n int
	cw := csv.NewWriter(w)
	wroteHeader := false
	for {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		row, ok := cr.rp.NextRow(ctx)
		if !ok {
			break
		}
		if !wroteHeader {
			if err := cw.Write(makeCSVHeaderFromRow(row)); err != nil {
				return 0, err
			}
			wroteHeader = true
		}
		if err := cw.Write(makeCSVRecordFromRow(row)); err != nil {
			return 0, err
		}
		cw.Flush()
	}
	return n, cw.Error()
}

func makeCSVHeaderFromRow(row Row) []string {
	hd := make([]string, len(row))
	for i, c := range row {
		hd[i] = c.Name
	}
	return hd
}

func makeCSVRecordFromRow(row Row) []string {
	r := make([]string, len(row))
	for i, c := range row {
		r[i] = strings.ReplaceAll(c.Text(), "\n", " ")
	}
	return r
}
