package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type JSONResult struct {
	rp RowProducer
}

func NewJSONResult(rp RowProducer) *JSONResult {
	return &JSONResult{rp: rp}
}

func (jr *JSONResult) Serialize(ctx context.Context, w io.Writer) (int, error) {
	var n int
	cr := []byte{'\n'}
	for {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		row, ok := jr.rp.NextRow(ctx)
		if !ok {
			return n, nil
		}
		m := make(map[string]any, len(row))
		for _, col := range row {
			v, err := col.JSONValue()
			if err != nil {
				continue
			}
			m[col.Name] = v
		}
		b, err := json.Marshal(m)
		if err != nil {
			return 0, fmt.Errorf("json marshalling result: %w", err)
		}
		wn, err := w.Write(b)
		if err != nil {
			return 0, fmt.Errorf("serializing result: %w", err)
		}
		wn, err = w.Write(cr)
		if err != nil {
			return 0, fmt.Errorf("serializing result: %w", err)
		}
		n += wn
	}
}
