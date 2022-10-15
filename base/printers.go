package base

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type JSONPrinter struct{}

func (pr JSONPrinter) Print(w io.Writer, rows []output.Row) error {
	result := output.NewSimpleRows(rows)
	jr := output.NewJSONResult(result)
	_, err := jr.Serialize(context.Background(), w)
	return err
}

type TablePrinter struct{}

func (pr TablePrinter) Print(w io.Writer, rows []output.Row) error {
	hd := map[string]struct{}{}
	drows := make([]map[string]output.Column, len(rows))
	for i, row := range rows {
		newRow := map[string]output.Column{}
		for ci, col := range row {
			// do not break the key column
			if ci == 0 {
				hd[col.Name] = struct{}{}
				newRow[col.Name] = col
				continue
			}
			// break out only complex types
			if col.Type == serialization.TypeJSONSerialization || col.Type == serialization.TypePortable {
				if col.Value != output.ValueNotDecoded {
					nc, err := col.RowExtensions()
					if err == nil {
						for _, sc := range nc {
							if sc.Name == "" {
								sc.Name = col.Name
							} else {
								sc.Name = fmt.Sprintf("%s.%s", col.Name, sc.Name)
							}
							hd[sc.Name] = struct{}{}
							newRow[sc.Name] = sc
							newRow[col.Name] = output.NewSkipColumn()
						}
						continue
					}
					newRow[col.Name] = output.Column{
						Type:  serialization.TypeString,
						Value: output.ValueNotDecoded,
					}
					continue
				}
			}
			hd[col.Name] = struct{}{}
			newRow[col.Name] = col
		}
		drows[i] = newRow
	}
	stdNames := []string{output.NameKey, output.NameKeyType, output.NameValue, output.NameValueType}
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
	nilCol := output.Column{Type: serialization.TypeNil}
	for ri, drow := range drows {
		row := make([]output.Column, len(header))
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
	result := output.NewSimpleRows(rows)
	// remove this. prefix from header cells
	thisPfx := fmt.Sprintf("%s.", output.NameValue)
	for i, h := range header {
		header[i] = strings.TrimPrefix(h, thisPfx)
	}
	tr := output.NewTableResult(header, result)
	_, err := tr.Serialize(context.Background(), os.Stdout)
	return err
}

func init() {
	plug.Registry.RegisterPrinter("json", &JSONPrinter{})
	plug.Registry.RegisterPrinter("table", &TablePrinter{})
}
