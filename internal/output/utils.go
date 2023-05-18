package output

import (
	"fmt"
	"math"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

func MakeTableFromRows(rows []Row, maxWidth int) (table.Row, []Row) {
	hd := NewOrderedSet[string]()
	drows := make([]map[string]Column, len(rows))
	for i, row := range rows {
		newRow := map[string]Column{}
		for _, col := range row {
			// do not break the key column
			// TODO: fix this properly
			if col.Name == NameKey {
				hd.Add(col.Name)
				newRow[col.Name] = col
				continue
			}
			// break out only complex types
			if canBreakOutValue(col) {
				// XXX: what if col.Value == ValueNotDecoded ?
				if col.Value != serialization.ValueNotDecoded {
					nc, err := col.RowExtensions()
					if err != nil {
						hd.Add(col.Name)
						newRow[col.Name] = Column{
							Type:  serialization.TypeString,
							Value: serialization.ValueNotDecoded,
						}
						continue
					}
					for _, sc := range nc {
						if sc.Name == "" {
							sc.Name = col.Name
						} else {
							sc.Name = fmt.Sprintf("%s.%s", col.Name, sc.Name)
						}
						hd.Add(sc.Name)
						newRow[sc.Name] = sc
					}
					hd.Add(col.Name)
					newRow[col.Name] = Column{Type: serialization.TypeSkip}
					continue
				}
			}
			hd.Add(col.Name)
			newRow[col.Name] = col
		}
		drows[i] = newRow
	}
	stdNames := []string{NameKey, NameKeyType, NameValue, NameValueType}
	var stdHeader []string
	for _, h := range stdNames {
		if hd.Contains(h) {
			stdHeader = append(stdHeader, h)
		}
	}
	// delete standard headers
	for _, h := range stdNames {
		hd.Delete(h)
	}
	header := hd.Items()
	header = append(stdHeader, header...)
	// create new rows
	nilCol := Column{Type: serialization.TypeNil}
	// width is the max width for this column
	width := make([]int, len(header))
	// max width cannot be smaller than the header
	for i, h := range header {
		width[i] = len(h)
	}
	align := make([]int, len(header))
	var alignSet bool
	for ri, drow := range drows {
		row := make([]Column, len(header))
		for i, h := range header {
			v, ok := drow[h]
			if !ok {
				v = nilCol
				v.Name = h
			}
			row[i] = v
			sv := v.Text()
			if len(sv) > width[i] {
				width[i] = len(sv)
			}
			if !alignSet {
				align[i] = alignmentForType(v.Type)
			}
		}
		alignSet = true
		rows[ri] = row
	}
	// remove this. prefix from header cells
	thisPfx := NameValue + "."
	columns := make(table.Row, len(header))
	for i, h := range header {
		columns[i] = table.Column{
			Header: strings.TrimPrefix(h, thisPfx),
			Align:  align[i] * minPositive(maxWidth, width[i]),
		}
	}
	return columns, rows
}

func minPositive[T ~int](a T, b T) T {
	if a <= 0 {
		if b > 0 {
			return b
		}
		return 0
	}
	if b <= 0 {
		if a > 0 {
			return a
		}
		return 0
	}
	if a < b {
		return a
	}
	return b
}

func canBreakOutValue(col Column) bool {
	return col.Type == serialization.TypeJSONSerialization ||
		col.Type == serialization.TypePortable ||
		col.Type == serialization.TypeCompact
}

func makeTableHeaderFromRow(row Row, maxWidth int) table.Row {
	dw := defaultWidthForFlexibleColumns(row, maxWidth)
	// the max width for a column is 1000, just an arbitrary decision --YT
	if dw > 1000 {
		dw = 1000
	}
	ww := 0
	for _, c := range row {
		ww += widthForType(c.Type, len(c.Name), dw)
	}
	hd := make(table.Row, len(row))
	for i, c := range row {
		hd[i] = table.Column{
			Header: c.Name,
			Align:  alignmentForType(c.Type) * widthForType(c.Type, len(c.Name), dw),
		}
	}
	return hd
}

func alignmentForType(t int32) int {
	switch t {
	case serialization.TypeByte, serialization.TypeInt8, serialization.TypeUInt16,
		serialization.TypeInt16, serialization.TypeInt32, serialization.TypeInt64,
		serialization.TypeFloat32, serialization.TypeFloat64,
		serialization.TypeJavaBigInteger, serialization.TypeJavaDecimal:
		return table.AlignRight
	default:
		return table.AlignLeft
	}
}

func widthForType(t int32, headerWidth, defaultWidth int) int {
	switch t {
	case serialization.TypeByte:
		return maxInt(headerWidth, len(fmt.Sprint(math.MaxInt8)))
	case serialization.TypeUInt16, serialization.TypeInt16:
		return maxInt(headerWidth, len(fmt.Sprint(math.MaxInt16)))
	case serialization.TypeInt32:
		return maxInt(headerWidth, len(fmt.Sprint(math.MaxInt32)))
	case serialization.TypeInt64:
		return maxInt(headerWidth, len(fmt.Sprint(math.MaxInt)))
	case serialization.TypeFloat32:
		return maxInt(headerWidth, len(fmt.Sprint(math.MaxFloat32)))
	case serialization.TypeFloat64:
		return maxInt(headerWidth, len(fmt.Sprint(math.MaxFloat64)))
	case serialization.TypeBool:
		return maxInt(headerWidth, len(fmt.Sprint(false)))
	case serialization.TypeUUID:
		return maxInt(headerWidth, len(fmt.Sprint(types.UUID{})))
	default:
		return maxInt(headerWidth, defaultWidth)
	}
}

func defaultWidthForFlexibleColumns(header Row, maxWidth int) int {
	if maxWidth < 1 {
		return 1
	}
	const padding = 3
	fixedSize := 0
	flexibleCount := 0
	for _, h := range header {
		switch h.Type {
		case serialization.TypeByte, serialization.TypeUInt16, serialization.TypeInt16,
			serialization.TypeInt32, serialization.TypeInt64,
			serialization.TypeFloat32, serialization.TypeFloat64,
			serialization.TypeBool, serialization.TypeUUID:
			fixedSize += widthForType(h.Type, len(h.Name), len(h.Name)) + padding
		default:
			flexibleCount++
		}
	}
	if flexibleCount == 0 {
		return 1
	}
	w := (maxWidth-fixedSize)/flexibleCount - padding
	if w < 0 {
		return 1
	}
	return w
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampInt(num, a, b int) int {
	return minInt(a, maxInt(num, b))
}
