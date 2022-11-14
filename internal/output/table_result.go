package output

import (
	"context"
	"fmt"
	"io"

	iserialization "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

type TableOutputMode int

const (
	TableOutModeDefault TableOutputMode = iota
	TableOutputModeCSV
	TableOutputModeHTML
	TableOutputModeMarkDown
)

type TableResult struct {
	header []string
	rp     RowProducer
}

func NewTableResult(header []string, rp RowProducer) *TableResult {
	return &TableResult{
		header: header,
		rp:     rp,
	}
}

func (tr *TableResult) Serialize(ctx context.Context, w io.Writer, mode TableOutputMode) (int, error) {
	var n int
	t := table.New(table.Config{Stdout: w})
	if tr.header != nil {
		cs := make([]table.Column, len(tr.header))
		for i, h := range tr.header {
			cs[i] = table.Column{
				Header: h,
				Align:  10,
			}
		}
		t.WriteHeader(cs)
	}
	for {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		vr, ok := tr.rp.NextRow()
		if !ok {
			break
		}
		row := make([]string, len(vr))
		for i, v := range vr {
			row[i] = fmt.Sprint(tr.convertColumn(v))
		}
		t.WriteRow(row)
	}
	return n, nil
}

func (tr *TableResult) convertColumn(col Column) any {
	switch col.Type {
	case iserialization.TypeByte, iserialization.TypeBool, iserialization.TypeUInt16,
		iserialization.TypeInt16, iserialization.TypeInt32, iserialization.TypeInt64,
		iserialization.TypeFloat32, iserialization.TypeFloat64, iserialization.TypeString:
		return col.Value
	case iserialization.TypeNil:
		return ValueNil
	case iserialization.TypeUnknown:
		return ValueUnknown
	case iserialization.TypeSkip:
		return ValueSkip
	default:
		return col.SingleLine()
	}
}
