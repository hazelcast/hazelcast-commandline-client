package serialization

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type Column struct {
	Name  string
	Type  int32
	Value any
}

func (co Column) Text() (s string) {
	if sl, ok := co.Value.(Texter); ok {
		return sl.Text()
	}
	if _, ok := co.Value.(NondecodedType); ok {
		return ValueNotDecoded
	}
	str, ok := ValueToText[co.Type]
	if !ok {
		return ValueNotDecoded
	}
	return str(co.Value)
}

func (co Column) RowExtensions() ([]Column, error) {
	switch co.Type {
	case TypeJSONSerialization:
		value := []byte(co.Value.(serialization.JSON))
		var m any
		if err := json.Unmarshal(value, &m); err != nil {
			return nil, errors.ErrNotDecoded
		}
		// TODO: nested fields
		return jsonValueToColumns(m), nil
	case TypePortable:
		value, ok := co.Value.(*GenericPortable)
		if !ok {
			return nil, errors.ErrNotDecoded
		}
		cols := make([]Column, len(value.Fields))
		for i, f := range value.Fields {
			cols[i] = Column{
				Name:  f.Name,
				Type:  f.Type,
				Value: f.Value,
			}
		}
		return cols, nil
	case TypeCompact:
		if v, ok := co.Value.(ColumnList); ok {
			return v, nil
		}
		return nil, errors.ErrNotDecoded
	}
	return ColumnList{co}, nil
}

func (col Column) JSONValue() (any, error) {
	if check.IsNil(col.Value) {
		return nil, nil
	}
	if v, ok := col.Value.(JSONValuer); ok {
		return v.JSONValue()
	}
	switch col.Type {
	case TypeNil:
		return nil, nil
	case TypePortable, TypeCompact,
		TypeByte, TypeBool, TypeInt8, TypeUInt16,
		TypeInt16, TypeInt32, TypeInt64,
		TypeFloat32, TypeFloat64, TypeString,
		TypeByteArray, TypeBoolArray, TypeInt8Array, TypeUInt16Array,
		TypeInt16Array, TypeInt32Array, TypeInt64Array,
		TypeFloat32Array, TypeFloat64Array, TypeStringArray,
		TypeJavaArray, TypeJavaArrayList, TypeJavaLinkedList,
		TypeJSONSerialization:
		return col.Value, nil
	}
	return col.Text(), nil
}

type ColumnList []Column

func (cs ColumnList) Text() string {
	const delim = "; "
	var sb strings.Builder
	if len(cs) == 0 {
		return ""
	}
	sb.WriteString(cs[0].Name)
	sb.WriteString(":")
	sb.WriteString(cs[0].Text())
	for _, c := range cs[1:] {
		sb.WriteString(delim)
		sb.WriteString(c.Name)
		sb.WriteString(":")
		sb.WriteString(c.Text())
	}
	return sb.String()
}

func (cs ColumnList) JSONValue() (any, error) {
	m := make(map[string]any, len(cs))
	for _, c := range cs {
		v, err := c.JSONValue()
		if err != nil {
			v = ValueNotDecoded
		}
		m[c.Name] = v
	}
	return m, nil
}

func jsonValueToColumns(value any) []Column {
	if vv, ok := value.(map[string]any); ok {
		cols := make([]Column, 0, len(vv))
		for k, v := range vv {
			cols = append(cols, jsonValueToColumn(k, v))
		}
		sort.Slice(cols, func(i, j int) bool {
			return cols[i].Name < cols[j].Name
		})
		return cols
	}
	return []Column{jsonValueToColumn("", value)}
}

func jsonValueToColumn(k string, value any) Column {
	if _, ok := value.(map[string]any); ok {
		// TODO: nested maps are not handled yet
		return Column{
			Name: k,
			Type: TypeNotDecoded,
		}
	}
	return Column{
		Name:  k,
		Type:  jsonValueToTypeID(value),
		Value: value,
	}
}
