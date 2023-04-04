package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"
)

type NondecodedType string

type JSONValuer interface {
	JSONValue() (any, error)
}

type Texter interface {
	Text() string
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

func ptrStringer[T any]() Stringer {
	return func(v any) string {
		if vv, ok := v.(T); ok {
			return fmt.Sprint(vv)
		}
		if vv, ok := v.(*T); ok {
			if vv == nil {
				return ValueNil
			}
			return fmt.Sprint(*vv)
		}
		return fmt.Sprintf("??%v", reflect.TypeOf(v))
	}
}

func sprintNilStringer[T any](v *T) string {
	if v == (*T)(nil) {
		return ValueNil
	}
	return fmt.Sprint(*v)
}

func arrayStringer[T any](v any) string {
	vv, ok := v.([]T)
	if !ok {
		if vv, ok := v.([]*T); ok {
			return arrayPtrStringer[T](vv)
		}
		return fmt.Sprintf("??%v", reflect.TypeOf(v))
	}
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

func arrayPtrStringer[T any](v any) string {
	vv := v.([]*T)
	var sb strings.Builder
	sb.WriteString("[")
	if len(vv) > 0 {
		sb.WriteString(sprintNilStringer(vv[0]))
		for _, x := range vv[1:] {
			sb.WriteString(", ")
			sb.WriteString(sprintNilStringer(x))
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
	return "??datetime"
}

func portableSerializer(v any) string {
	return ""
}

var ValueToText = map[int32]Stringer{
	TypeNil: staticStringer(ValueNil),

	//TypePortable
	// TypeDataSerializable

	TypeByte:         ptrStringer[byte](),
	TypeBool:         ptrStringer[bool](),
	TypeUInt16:       ptrStringer[uint16](),
	TypeInt16:        ptrStringer[int16](),
	TypeInt32:        ptrStringer[int32](),
	TypeInt64:        ptrStringer[int64](),
	TypeFloat32:      ptrStringer[float32](),
	TypeFloat64:      ptrStringer[float64](),
	TypeString:       ptrStringer[string](),
	TypeByteArray:    arrayStringer[uint8],
	TypeBoolArray:    arrayStringer[bool],
	TypeUInt16Array:  arrayStringer[uint16],
	TypeInt16Array:   arrayStringer[int16],
	TypeInt32Array:   arrayStringer[int32],
	TypeInt64Array:   arrayStringer[int64],
	TypeFloat32Array: arrayStringer[float32],
	TypeFloat64Array: arrayStringer[float64],
	TypeStringArray:  arrayStringer[string],
	TypeUUID:         sprintStringer,

	// TypeSimpleEntry
	// TypeSimpleImmutableEntry

	TypeJavaClass:      javaClassStringer,
	TypeJavaDate:       timeStringer,
	TypeJavaBigInteger: sprintStringer,
	TypeJavaDecimal:    ptrStringer[types.Decimal](),
	/*
		TypeJavaArray:                            "JAVA_ARRAY",
		TypeJavaArrayList:                        "JAVA_ARRAY_LIST",
		TypeJavaLinkedList:                       "JAVA_LINKED_LIST",

	*/
	TypeJavaLocalDate:      timeStringer,
	TypeJavaLocalTime:      timeStringer,
	TypeJavaLocalDateTime:  timeStringer,
	TypeJavaOffsetDateTime: timeStringer,

	//

	TypeInt8:      ptrStringer[int8](),
	TypeInt8Array: arrayStringer[int8],
}

//func MakeText()
