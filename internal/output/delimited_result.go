package output

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type DelimitedResult struct {
	delim      string
	rp         RowProducer
	singleLine bool
}

func NewDelimitedResult(delim string, rp RowProducer, oneline bool) *DelimitedResult {
	return &DelimitedResult{
		delim:      delim,
		rp:         rp,
		singleLine: oneline,
	}
}

func (d DelimitedResult) Serialize(ctx context.Context, w io.Writer) (int, error) {
	var sb strings.Builder
	var n int
	for {
		if ctx.Err() != nil {
			return n, ctx.Err()
		}
		row, ok := d.rp.NextRow(ctx)
		if !ok {
			return n, nil
		}
		if len(row) == 0 {
			continue
		}
		sb.WriteString(d.adapt(row[0]))
		for _, r := range row[1:] {
			sb.WriteString(d.delim)
			sb.WriteString(d.adapt(r))
		}
		sb.WriteString("\n")
		wn, err := w.Write([]byte(sb.String()))
		if err != nil {
			return 0, fmt.Errorf("serializing result: %w", err)
		}
		n += wn
		sb.Reset()
	}
}

func (d DelimitedResult) adapt(col Column) string {
	text := col.Text()
	if !d.singleLine {
		return text
	}
	return strings.ReplaceAll(text, "\n", " ")
}
