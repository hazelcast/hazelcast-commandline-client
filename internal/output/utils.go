package output

import (
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/table"
)

func MakeTable(rows []Row) ([]table.Column, []Row) {
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
			if col.Type == serialization.TypeJSONSerialization || col.Type == serialization.TypePortable || col.Type == serialization.TypeCompact {
				// XXX: what if col.Value == ValueNotDecoded ?
				if col.Value != ValueNotDecoded {
					nc, err := col.RowExtensions()
					if err != nil {
						hd.Add(col.Name)
						newRow[col.Name] = Column{
							Type:  serialization.TypeString,
							Value: ValueNotDecoded,
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
						newRow[col.Name] = NewSkipColumn()
					}
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
			sv := fmt.Sprint(v.Value)
			if len(sv) > width[i] {
				width[i] = len(sv)
			}
			if !alignSet {
				switch v.Type {
				case serialization.TypeByte, serialization.TypeUInt16, serialization.TypeInt16,
					serialization.TypeInt32, serialization.TypeInt64,
					serialization.TypeFloat32, serialization.TypeFloat64,
					serialization.TypeJavaBigInteger, serialization.TypeJavaDecimal:
					align[i] = -1
				default:
					align[i] = 1
				}
			}
		}
		alignSet = true
		rows[ri] = row
	}
	// remove this. prefix from header cells
	thisPfx := fmt.Sprintf("%s.", NameValue)
	columns := make([]table.Column, len(header))
	for i, h := range header {
		columns[i] = table.Column{
			Header: strings.TrimPrefix(h, thisPfx),
			Align:  align[i] * width[i],
		}
	}
	return columns, rows
}
