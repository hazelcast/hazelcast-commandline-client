package output

import (
	"context"
	"fmt"
	"io"

	"github.com/fatih/color"

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
	header []table.Column
	rp     RowProducer
}

func NewTableResult(header []table.Column, rp RowProducer) *TableResult {
	return &TableResult{
		header: header,
		rp:     rp,
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
			t.WriteHeader(MakeHeaderFromRow(vr))
			wroteHeader = true
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
