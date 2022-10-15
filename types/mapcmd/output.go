package mapcmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	hazelcast "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

func printSingleValue(value any, valueType int32, showType bool, ot output.Type) error {
	row := output.Row{{Name: output.NameValue, Type: valueType, Value: value}}
	if showType {
		row = append([]output.Column{
			{
				Name:  output.NameKeyType,
				Type:  serialization.TypeString,
				Value: serialization.TypeToString(valueType),
			},
		}, row...)
	}
	rows := []output.Row{row}
	result := output.NewSimpleRows(rows)
	if ot == output.TypeDelimited {
		dr := output.NewDelimitedResult("\n", result, false)
		// ignoring the error
		_, err := dr.Serialize(context.Background(), os.Stdout)
		return err
	}
	if ot == output.TypeJSON {
		jr := output.NewJSONResult(result)
		_, err := jr.Serialize(context.Background(), os.Stdout)
		return err
	}
	panic(fmt.Errorf("unsupported output type: %d", ot))
}

func decodePairs(ic *hazelcast.ClientInternal, pairs []hazelcast.Pair, showType bool) []output.Row {
	rows := make([]output.Row, 0, len(pairs))
	vs := []any{}
	for _, pair := range pairs {
		row := make(output.Row, 0, 4)
		kt, key := ensureTypeValue(ic, pair.Key.(hazelcast.Data))
		row = append(row, output.NewKeyColumn(kt, key))
		if showType {
			row = append(row, output.NewKeyTypeColumn(kt))
		}
		vt, value := ensureTypeValue(ic, pair.Value.(hazelcast.Data))
		vs = append(vs, value)
		row = append(row, output.NewValueColumn(vt, value))
		if showType {
			row = append(row, output.NewValueTypeColumn(vt))
		}
		rows = append(rows, row)
	}
	return rows
}

func printPairs(ic *hazelcast.ClientInternal, pairs []hazelcast.Pair, showType bool, ot output.Type) error {
	return printRows(decodePairs(ic, pairs, showType), ot)
}

func printRows(rows []output.Row, ot output.Type) error {
	if ot == output.TypeDelimited {
		result := output.NewSimpleRows(rows)
		dr := output.NewDelimitedResult("\t", result, true)
		_, err := dr.Serialize(context.Background(), os.Stdout)
		return err
	}
	if ot == output.TypeJSON {
		result := output.NewSimpleRows(rows)
		jr := output.NewJSONResult(result)
		_, err := jr.Serialize(context.Background(), os.Stdout)
		return err
	}
	if ot == output.TypeTable {
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
		stdHeader := []string{output.NameKey}
		if _, ok := hd[output.NameKeyType]; ok {
			stdHeader = append(stdHeader, output.NameKeyType)
		}
		stdHeader = append(stdHeader, output.NameValue)
		if _, ok := hd[output.NameValueType]; ok {
			stdHeader = append(stdHeader, output.NameValueType)
		}
		header := make([]string, 0, len(hd))
		// delete standard headers
		delete(hd, output.NameKeyType)
		delete(hd, output.NameKey)
		delete(hd, output.NameValueType)
		delete(hd, output.NameValue)
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
	panic(fmt.Errorf("unsupported output type: %d", ot))
}

func ensureTypeValue(ic *hazelcast.ClientInternal, data hazelcast.Data) (int32, any) {
	t := data.Type()
	v, err := ic.DecodeData(data)
	if err != nil {
		v = serialization.NondecodedType(serialization.TypeToString(t))
	}
	return t, v
}
