package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func MakeTable(rows []Row) ([]string, []Row) {
	hd := map[string]struct{}{}
	drows := make([]map[string]Column, len(rows))
	for i, row := range rows {
		newRow := map[string]Column{}
		for _, col := range row {
			// do not break the key column
			// TODO: fix this properly
			if col.Name == NameKey {
				hd[col.Name] = struct{}{}
				newRow[col.Name] = col
				continue
			}
			// break out only complex types
			if col.Type == serialization.TypeJSONSerialization || col.Type == serialization.TypePortable || col.Type == serialization.TypeCompact {
				// XXX: what if col.Value == ValueNotDecoded ?
				if col.Value != ValueNotDecoded {
					nc, err := col.RowExtensions()
					if err != nil {
						hd[col.Name] = struct{}{}
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
						hd[sc.Name] = struct{}{}
						newRow[sc.Name] = sc
						newRow[col.Name] = NewSkipColumn()
					}
					continue
				}
			}
			hd[col.Name] = struct{}{}
			newRow[col.Name] = col
		}
		drows[i] = newRow
	}
	stdNames := []string{NameKey, NameKeyType, NameValue, NameValueType}
	var stdHeader []string
	for _, h := range stdNames {
		if _, ok := hd[h]; ok {
			stdHeader = append(stdHeader, h)
		}
	}
	header := make([]string, 0, len(hd))
	// delete standard headers
	for _, h := range stdNames {
		delete(hd, h)
	}
	for h := range hd {
		header = append(header, h)
	}
	sort.Slice(header, func(i, j int) bool {
		return header[i] < header[j]
	})
	header = append(stdHeader, header...)
	// create new rows
	nilCol := Column{Type: serialization.TypeNil}
	for ri, drow := range drows {
		row := make([]Column, len(header))
		for i, h := range header {
			v, ok := drow[h]
			if !ok {
				v = nilCol
				v.Name = h
			}
			row[i] = v
		}
		rows[ri] = row
	}
	// remove this. prefix from header cells
	thisPfx := fmt.Sprintf("%s.", NameValue)
	for i, h := range header {
		header[i] = strings.TrimPrefix(h, thisPfx)
	}
	return header, rows
}
