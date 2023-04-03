package serialization

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/errors"
)

type NondecodedType string

type SingleLiner interface {
	SingleLine() string
}

type MultiLiner interface {
	MultiLine() []string
}

type JSONValuer interface {
	JSONValue() (any, error)
}

type Column struct {
	Name  string
	Type  int32
	Value any
}

func (co Column) SingleLine() (s string) {
	if sl, ok := co.Value.(SingleLiner); ok {
		return sl.SingleLine()
	}
	if _, ok := co.Value.(NondecodedType); ok {
		return ValueNotDecoded
	}
	str, ok := ValueToText[co.Type]
	if ok {
		return str(co.Value)
	}
	switch co.Type {
	case TypeNil:
		s = ValueNil
	case TypePortable:
		s = co.Value.(*GenericPortable).String()
	case TypeDataSerializable:
		s = ValueNotDecoded
	case TypeByte, TypeBool, TypeUInt16,
		TypeInt16, TypeInt32, TypeInt64,
		TypeFloat32, TypeFloat64, TypeString,
		TypeByteArray, TypeBoolArray, TypeUInt16Array,
		TypeInt16Array, TypeInt32Array, TypeInt64Array,
		TypeFloat32Array, TypeFloat64Array, TypeStringArray:
		s = fmt.Sprint(co.Value)
	case TypeUUID:
		s = co.Value.(types.UUID).String()
	case TypeSimpleEntry, TypeSimpleImmutableEntry:
		s = ValueNotDecoded
	case TypeJavaClass:
		s = co.Value.(string)
	case TypeJavaDate:
		s = co.Value.(time.Time).Format(time.RFC3339)
	case TypeJavaBigInteger:
		s = co.Value.(*big.Int).String()
	case TypeJavaArray, TypeJavaArrayList, TypeJavaLinkedList:
		s = fmt.Sprint(co.Value)
	case TypeJavaDefaultTypeCopyOnWriteArrayList, TypeJavaDefaultTypeHashMap,
		TypeJavaDefaultTypeConcurrentSkipListMap, TypeJavaDefaultTypeConcurrentHashMap,
		TypeJavaDefaultTypeLinkedHashMap, TypeJavaDefaultTypeTreeMap,
		TypeJavaDefaultTypeHashSet, TypeJavaDefaultTypeTreeSet,
		TypeJavaDefaultTypeLinkedHashSet, TypeJavaDefaultTypeCopyOnWriteArraySet,
		TypeJavaDefaultTypeConcurrentSkipListSet, TypeJavaDefaultTypeArrayDeque,
		TypeJavaDefaultTypeLinkedBlockingQueue, TypeJavaDefaultTypeArrayBlockingQueue,
		TypeJavaDefaultTypePriorityBlockingQueue, TypeJavaDefaultTypeDelayQueue,
		TypeJavaDefaultTypeSynchronousQueue, TypeJavaDefaultTypeLinkedTransferQueue,
		TypeJavaDefaultTypePriorityQueue, TypeJavaDefaultTypeOptional:
		s = ValueNotDecoded
	case TypeJavaDecimal:
		sr, err := MarshalDecimal(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else {
			s = sr
		}
	case TypeJavaLocalDate:
		sr, err := MarshalLocalDate(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else if sr == nil {
			s = ValueNil
		} else {
			s = *sr
		}
	case TypeJavaLocalTime:
		sr, err := MarshalLocalTime(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else if sr == nil {
			s = ValueNil
		} else {
			s = *sr
		}
	case TypeJavaLocalDateTime:
		sr, err := MarshalLocalDateTime(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else if sr == nil {
			s = ValueNil
		} else {
			s = *sr
		}
	case TypeJavaOffsetDateTime:
		sr, err := MarshalOffsetDateTime(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else if sr == nil {
			s = ValueNil
		} else {
			s = *sr
		}
	case TypeCompact:
		sr, err := structToString(co.Value)
		if err != nil {
			sr = ValueNotDecoded
		}
		s = sr
	case TypeCompactWithSchema, TypeJavaDefaultTypeSerializable,
		TypeJavaDefaultTypeExternalizable, TypeCsharpCLRSerializationType,
		TypePythonPickleSerializationType:
		s = ValueNotDecoded
	case TypeJSONSerialization:
		sp := strings.Split(string(co.Value.(serialization.JSON)), "\n")
		for i, line := range sp {
			sp[i] = strings.TrimSpace(line)
		}
		s = strings.Join(sp, "")
	case TypeGobSerialization:
		s = fmt.Sprint(co.Value)
	case TypeHibernate3TypeHibernateCacheKey, TypeHibernate3TypeHibernateCacheEntry,
		TypeHibernate4TypeHibernateCacheKey, TypeHibernate4TypeHibernateCacheEntry,
		TypeHibernate5TypeHibernateCacheKey, TypeHibernate5TypeHibernateCacheEntry,
		TypeHibernate5TypeHibernateNaturalIDKey,
		TypeJetSerializerFirst, TypeJetSerializerLast:
		s = ValueNotDecoded
	case TypeUnknown:
		s = ValueUnknown
	case TypeSkip:
		s = ValueSkip
	case TypeNotDecoded:
		s = ValueNotDecoded
	default:
		s = ValueUnknown
	}
	idx := strings.Index(s, "\n")
	if idx < 0 {
		idx = len(s)
	}
	return s[:idx]
}

func (co Column) MultiLine() string {
	switch co.Type {
	case TypeNil:
		return ValueSkip
	case TypeUnknown:
		return ValueUnknown
	}
	return fmt.Sprint(co.Value)
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
				Type:  f.Type.ToTypeID(),
				Value: f.Value,
			}
		}
		return cols, nil
	case TypeCompact:
		value, err := structToString(co.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		}
		// the same code path with JSON
		var m any
		if err := json.Unmarshal([]byte(value), &m); err != nil {
			return nil, errors.ErrNotDecoded
		}
		// TODO: nested fields
		return jsonValueToColumns(m), nil
	}
	return []Column{co}, nil
}

func (col Column) JSONValue() (any, error) {
	if v, ok := col.Value.(JSONValuer); ok {
		return v.JSONValue()
	}
	switch col.Type {
	case TypeNil:
		return nil, nil
	case TypePortable, TypeCompact:
		return col.Value, nil
	case TypeDataSerializable:
		return nil, errors.ErrNotDecoded
	case TypeByte, TypeBool, TypeUInt16,
		TypeInt16, TypeInt32, TypeInt64,
		TypeFloat32, TypeFloat64, TypeString,
		TypeByteArray, TypeBoolArray, TypeUInt16Array,
		TypeInt16Array, TypeInt32Array, TypeInt64Array,
		TypeFloat32Array, TypeFloat64Array, TypeStringArray:
		return col.Value, nil
	case TypeUUID:
		return col.Value.(types.UUID).String(), nil
	case TypeSimpleEntry, TypeSimpleImmutableEntry:
		return nil, errors.ErrNotDecoded
	case TypeJavaClass:
		return col.Value.(string), nil
	case TypeJavaDate:
		return col.Value.(time.Time).Format(time.RFC3339), nil
	case TypeJavaBigInteger:
		return col.Value.(*big.Int).String(), nil
	case TypeJavaDecimal:
		return MarshalDecimal(col.Value)
	case TypeJavaArray, TypeJavaArrayList, TypeJavaLinkedList:
		return col.Value, nil
	case TypeJavaLocalDate:
		sr, err := MarshalLocalDate(col.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		} else if sr == nil {
			return nil, nil
		} else {
			return *sr, nil
		}
	case TypeJavaLocalTime:
		sr, err := MarshalLocalTime(col.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		} else if sr == nil {
			return nil, nil
		} else {
			return *sr, nil
		}
	case TypeJavaLocalDateTime:
		sr, err := MarshalLocalDateTime(col.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		} else if sr == nil {
			return nil, nil
		} else {
			return *sr, nil
		}
	case TypeJavaOffsetDateTime:
		sr, err := MarshalOffsetDateTime(col.Value)
		if err != nil {
			return nil, errors.ErrNotDecoded
		} else if sr == nil {
			return nil, nil
		} else {
			return *sr, nil
		}
	case TypeJSONSerialization:
		return col.Value, nil
	}
	if sl, ok := col.Value.(SingleLiner); ok {
		return sl.SingleLine(), nil
	}
	return nil, errors.ErrNotDecoded
}

type ColumnList []Column

func (cs ColumnList) SingleLine() string {
	const delim = "; "
	var sb strings.Builder
	if len(cs) == 0 {
		return ""
	}
	sb.WriteString(cs[0].Name)
	sb.WriteString(":")
	sb.WriteString(cs[0].SingleLine())
	for _, c := range cs[1:] {
		sb.WriteString(delim)
		sb.WriteString(c.Name)
		sb.WriteString(":")
		sb.WriteString(c.SingleLine())
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

func jsonValueToTypeID(v any) int32 {
	switch v.(type) {
	case []any:
		return TypeJavaArray
	case string:
		return TypeString
	case float64:
		return TypeFloat64
	case bool:
		return TypeBool
	case nil:
		return TypeNil
	}
	return TypeUnknown
}

func structToString(v any) (string, error) {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Pointer {
		if vv.IsNil() {
			return ValueNil, nil
		}
		vv = vv.Elem()
	}
	if vv.Kind() == reflect.Struct {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return fmt.Sprintf("%v", v), nil
}

type Stringer func(any) string

func staticStringer(v string) Stringer {
	return func(a any) string {
		return v
	}
}

func sprintStringer(v any) string {
	if v == nil {
		return ValueNil
	}
	return fmt.Sprint(v)
}

func arrayStringer[T any](v any) string {
	vv := v.([]T)
	var sb strings.Builder
	sb.WriteString("[")
	if len(vv) > 0 {
		sb.WriteString(fmt.Sprint(vv[0]))
		for _, x := range vv[1:] {
			sb.WriteString(", ")
			sb.WriteString(fmt.Sprint(x))
		}
	}
	sb.WriteString("]")
	return sb.String()
}

func javaClassStringer(v any) string {
	return fmt.Sprintf("Java Class: %v", v)
}

func timeStringer(v any) string {
	switch vv := v.(type) {
	case time.Time:
		return vv.Format(time.RFC3339)
	case *time.Time:
		return vv.Format(time.RFC3339)
	case types.LocalTime:
		return time.Time(vv).Format("15:04:05")
	case *types.LocalTime:
		if vv == nil {
			return ValueNil
		}
		return (*time.Time)(vv).Format("15:04:05")
	case types.LocalDate:
		return time.Time(vv).Format("2006-01-02")
	case *types.LocalDate:
		if vv == nil {
			return ValueNil
		}
		return (*time.Time)(vv).Format("2006-01-02")
	case types.LocalDateTime:
		return time.Time(vv).Format("2006-01-02 15:04:05")
	case *types.LocalDateTime:
		if vv == nil {
			return ValueNil
		}
		return (*time.Time)(vv).Format("2006-01-02 15:04:05")
	case types.OffsetDateTime:
		return time.Time(vv).Format(time.RFC3339)
	case *types.OffsetDateTime:
		if vv == nil {
			return ValueNil
		}
		return (*time.Time)(vv).Format(time.RFC3339)
	}
	return "UNHANDLED_DATE_TIME"
}

var ValueToText = map[int32]Stringer{
	TypeNil:            staticStringer(ValueNil),
	TypeByte:           sprintStringer,
	TypeBool:           sprintStringer,
	TypeUInt16:         sprintStringer,
	TypeInt16:          sprintStringer,
	TypeInt32:          sprintStringer,
	TypeInt64:          sprintStringer,
	TypeFloat32:        sprintStringer,
	TypeFloat64:        sprintStringer,
	TypeString:         sprintStringer,
	TypeByteArray:      arrayStringer[byte],
	TypeBoolArray:      arrayStringer[bool],
	TypeUInt16Array:    arrayStringer[uint16],
	TypeInt16Array:     arrayStringer[int16],
	TypeInt32Array:     arrayStringer[int32],
	TypeInt64Array:     arrayStringer[int64],
	TypeFloat32Array:   arrayStringer[float32],
	TypeFloat64Array:   arrayStringer[float64],
	TypeStringArray:    arrayStringer[string],
	TypeUUID:           sprintStringer,
	TypeJavaClass:      javaClassStringer,
	TypeJavaDate:       timeStringer,
	TypeJavaBigInteger: sprintStringer,
	TypeJavaDecimal:    sprintStringer,
	/*
		TypeJavaArray:                            "JAVA_ARRAY",
		TypeJavaArrayList:                        "JAVA_ARRAY_LIST",
		TypeJavaLinkedList:                       "JAVA_LINKED_LIST",

	*/
}

//func MakeText()
