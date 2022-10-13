package output

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/errors"
	iserialization "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type Column struct {
	Name  string
	Type  int32
	Value any
}

func NewStringColumn(value string) Column {
	return Column{Type: iserialization.TypeString, Value: value}
}

func NewNilColumn() Column {
	return Column{Type: iserialization.TypeNil}
}

func NewSkipColumn() Column {
	return Column{Type: iserialization.TypeSkip}
}

func NewKeyColumn(kt int32, key any) Column {
	return Column{
		Name:  NameKey,
		Type:  kt,
		Value: key,
	}
}

func NewKeyTypeColumn(kt int32) Column {
	return Column{
		Name:  NameKeyType,
		Type:  iserialization.TypeString,
		Value: iserialization.TypeToString(kt),
	}
}

func NewValueColumn(vt int32, value any) Column {
	return Column{
		Name:  NameValue,
		Type:  vt,
		Value: value,
	}
}

func NewValueTypeColumn(vt int32) Column {
	return Column{
		Name:  NameValueType,
		Type:  iserialization.TypeString,
		Value: iserialization.TypeToString(vt),
	}
}

func (co Column) SingleLine() (s string) {
	/*
		defer func() {
			if e := recover(); e != nil {
				s = ValueNotDecoded
			}
		}()
	*/
	if sl, ok := co.Value.(SingleLiner); ok {
		return sl.SingleLine()
	}
	if _, ok := co.Value.(iserialization.NondecodedType); ok {
		return ValueNotDecoded
	}
	switch co.Type {
	case iserialization.TypeNil:
		s = ValueNil
	case iserialization.TypePortable:
		s = co.Value.(*iserialization.GenericPortable).String()
	case iserialization.TypeDataSerializable:
		s = ValueNotDecoded
	case iserialization.TypeByte, iserialization.TypeBool, iserialization.TypeUInt16,
		iserialization.TypeInt16, iserialization.TypeInt32, iserialization.TypeInt64,
		iserialization.TypeFloat32, iserialization.TypeFloat64, iserialization.TypeString,
		iserialization.TypeByteArray, iserialization.TypeBoolArray, iserialization.TypeUInt16Array,
		iserialization.TypeInt16Array, iserialization.TypeInt32Array, iserialization.TypeInt64Array,
		iserialization.TypeFloat32Array, iserialization.TypeFloat64Array:
		s = fmt.Sprintf("%v", co.Value)
	case iserialization.TypeStringArray:
		s = fmt.Sprintf("%v", co.Value)
	case iserialization.TypeUUID:
		s = co.Value.(types.UUID).String()
	case iserialization.TypeSimpleEntry, iserialization.TypeSimpleImmutableEntry:
		s = ValueNotDecoded
	case iserialization.TypeJavaClass:
		s = co.Value.(string)
	case iserialization.TypeJavaDate:
		s = co.Value.(time.Time).Format(time.RFC3339)
	case iserialization.TypeJavaBigInteger:
		s = co.Value.(*big.Int).String()
	case iserialization.TypeJavaDecimal:
		s = iserialization.MarshalDecimal(co.Value)
	case iserialization.TypeJavaArray, iserialization.TypeJavaArrayList, iserialization.TypeJavaLinkedList:
		s = fmt.Sprintf("%v", co.Value)
	case iserialization.TypeJavaDefaultTypeCopyOnWriteArrayList:
		fallthrough
	case iserialization.TypeJavaDefaultTypeHashMap:
		fallthrough
	case iserialization.TypeJavaDefaultTypeConcurrentSkipListMap:
		fallthrough
	case iserialization.TypeJavaDefaultTypeConcurrentHashMap:
		fallthrough
	case iserialization.TypeJavaDefaultTypeLinkedHashMap:
		fallthrough
	case iserialization.TypeJavaDefaultTypeTreeMap:
		fallthrough
	case iserialization.TypeJavaDefaultTypeHashSet:
		fallthrough
	case iserialization.TypeJavaDefaultTypeTreeSet:
		fallthrough
	case iserialization.TypeJavaDefaultTypeLinkedHashSet:
		fallthrough
	case iserialization.TypeJavaDefaultTypeCopyOnWriteArraySet:
		fallthrough
	case iserialization.TypeJavaDefaultTypeConcurrentSkipListSet:
		fallthrough
	case iserialization.TypeJavaDefaultTypeArrayDeque:
		fallthrough
	case iserialization.TypeJavaDefaultTypeLinkedBlockingQueue:
		fallthrough
	case iserialization.TypeJavaDefaultTypeArrayBlockingQueue:
		fallthrough
	case iserialization.TypeJavaDefaultTypePriorityBlockingQueue:
		fallthrough
	case iserialization.TypeJavaDefaultTypeDelayQueue:
		fallthrough
	case iserialization.TypeJavaDefaultTypeSynchronousQueue:
		fallthrough
	case iserialization.TypeJavaDefaultTypeLinkedTransferQueue:
		fallthrough
	case iserialization.TypeJavaDefaultTypePriorityQueue:
		fallthrough
	case iserialization.TypeJavaDefaultTypeOptional:
		s = ValueNotDecoded
	case iserialization.TypeJavaLocalDate:
		sr, err := iserialization.MarshalLocalDate(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else if sr == nil {
			s = ValueNil
		} else {
			s = *sr
		}
	case iserialization.TypeJavaLocalTime:
		sr, err := iserialization.MarshalLocalTime(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else if sr == nil {
			s = ValueNil
		} else {
			s = *sr
		}
	case iserialization.TypeJavaLocalDateTime:
		sr, err := iserialization.MarshalLocalDateTime(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else if sr == nil {
			s = ValueNil
		} else {
			s = *sr
		}
	case iserialization.TypeJavaOffsetDateTime:
		sr, err := iserialization.MarshalOffsetDateTime(co.Value)
		if err != nil {
			s = ValueNotDecoded
		} else if sr == nil {
			s = ValueNil
		} else {
			s = *sr
		}
	case iserialization.TypeCompact:
		fallthrough
	case iserialization.TypeCompactWithSchema:
		fallthrough
	case iserialization.TypeJavaDefaultTypeSerializable:
		fallthrough
	case iserialization.TypeJavaDefaultTypeExternalizable:
		fallthrough
	case iserialization.TypeCsharpCLRSerializationType:
		fallthrough
	case iserialization.TypePythonPickleSerializationType:
		s = ValueNotDecoded
	case iserialization.TypeJSONSerialization:
		sp := strings.Split(string(co.Value.(serialization.JSON)), "\n")
		for i, line := range sp {
			sp[i] = strings.TrimSpace(line)
		}
		s = strings.Join(sp, "")
	case iserialization.TypeGobSerialization:
		s = fmt.Sprintf("%v", co.Value)
	case iserialization.TypeHibernate3TypeHibernateCacheKey:
		fallthrough
	case iserialization.TypeHibernate3TypeHibernateCacheEntry:
		fallthrough
	case iserialization.TypeHibernate4TypeHibernateCacheKey:
		fallthrough
	case iserialization.TypeHibernate4TypeHibernateCacheEntry:
		fallthrough
	case iserialization.TypeHibernate5TypeHibernateCacheKey:
		fallthrough
	case iserialization.TypeHibernate5TypeHibernateCacheEntry:
		fallthrough
	case iserialization.TypeHibernate5TypeHibernateNaturalIDKey:
		fallthrough
	case iserialization.TypeJetSerializerFirst:
		fallthrough
	case iserialization.TypeJetSerializerLast:
		s = ValueNotDecoded
	case iserialization.TypeUnknown:
		s = ValueUnknown
	case iserialization.TypeSkip:
		s = ValueSkip
	case iserialization.TypeNotDecoded:
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
	case iserialization.TypeNil:
		return ValueSkip
	case iserialization.TypeUnknown:
		return ValueUnknown
	}
	return fmt.Sprintf("%v", co.Value)
}

func (co Column) RowExtensions() ([]Column, error) {
	switch co.Type {
	case iserialization.TypeJSONSerialization:
		value := []byte(co.Value.(serialization.JSON))
		var m any
		if err := json.Unmarshal(value, &m); err != nil {
			return nil, errors.ErrNotDecoded
		}
		// TODO: nested fields
		return jsonValueToColumns(m), nil
	case iserialization.TypePortable:
		value, ok := co.Value.(*iserialization.GenericPortable)
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
	}
	return []Column{co}, nil
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
			Type: iserialization.TypeNotDecoded,
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
		return iserialization.TypeJavaArray
	case string:
		return iserialization.TypeString
	case float64:
		return iserialization.TypeFloat64
	case bool:
		return iserialization.TypeBool
	case nil:
		return iserialization.TypeNil
	}
	return iserialization.TypeUnknown
}
