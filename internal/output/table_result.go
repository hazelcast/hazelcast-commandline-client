package output

import (
	"context"
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	iserialization "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
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
	rp     RowProvider
}

func NewTableResult(header []string, rp RowProvider) *TableResult {
	return &TableResult{
		header: header,
		rp:     rp,
	}
}

func (tr *TableResult) Serialize(ctx context.Context, w io.Writer, mode TableOutputMode) (int, error) {
	var n int
	header := make(table.Row, len(tr.header))
	for i, h := range tr.header {
		header[i] = h
	}
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.Style().Format.Header = text.FormatDefault
	t.AppendHeader(header)
	for {
		if ctx.Err() != nil {
			return 0, nil
		}
		vr, ok := tr.rp.NextRow()
		if !ok {
			break
		}
		row := make(table.Row, len(vr))
		for i, v := range vr {
			row[i] = tr.convertColumn(v)
		}
		t.AppendRow(row)
	}
	switch mode {
	case TableOutputModeCSV:
		t.RenderCSV()
	case TableOutputModeHTML:
		t.RenderHTML()
	case TableOutputModeMarkDown:
		t.RenderMarkdown()
	default:
		t.Render()
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
