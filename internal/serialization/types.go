package serialization

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
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

type Stringer func(any) string

func staticStringer(v string) Stringer {
	return func(a any) string {
		return v
	}
}

func sprintStringer(v any) string {
	if check.IsNil(v) {
		return ValueNil
	}
	if vv, ok := v.(Texter); ok {
		return vv.Text()
	}
	return fmt.Sprint(v)
}

func ptrStringer[T any]() Stringer {
	return func(v any) string {
		if v == nil {
			return ValueNil
		}
		if vv, ok := v.(T); ok {
			return fmt.Sprint(vv)
		}
		if vv, ok := v.(*T); ok {
			if vv == nil {
				return ValueNil
			}
			return fmt.Sprint(*vv)
		}
		return fmt.Sprintf("?ptr:%v?", reflect.TypeOf(v))
	}
}

func sprintNilStringer[T any](v *T) string {
	if v == (*T)(nil) {
		return ValueNil
	}
	return fmt.Sprint(*v)
}

func simplify(v any) any {
	if vc, ok := v.(ColumnList); ok {
		return convertToAnyArray(vc)
	}
	if vc, ok := v.(ColumnMap); ok {
		return convertToAnyArray(vc)
	}
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Slice {
		l := vv.Len()
		a := make([]any, l)
		for i := 0; i < l; i++ {
			a[i] = vv.Index(i).Interface()
		}
		return a
	}
	return v
}

func convertToAnyArray[T any](v []T) []any {
	r := make([]any, len(v))
	for i, x := range v {
		r[i] = x
	}
	return r
}

func arrayStringer[T any](stringer Stringer) Stringer {
	return func(v any) string {
		vv, ok := v.([]T)
		if !ok {
			if vv, ok := v.([]*T); ok {
				return arrayPtrStringer[T](vv)
			}
			return fmt.Sprintf("?array:%v?", reflect.TypeOf(v))
		}
		var sb strings.Builder
		sb.WriteString("[")
		if len(vv) > 0 {
			sb.WriteString(stringer(vv[0]))
			for _, x := range vv[1:] {
				sb.WriteString(", ")
				sb.WriteString(stringer(x))
			}
		}
		sb.WriteString("]")
		return sb.String()
	}
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

func arrayAnyStringer[T any](stringer Stringer) Stringer {
	return func(v any) string {
		vv, ok := v.([]T)
		if !ok {
			vv, ok := v.([]*T)
			if !ok {
				return "?array?"
			}
			return arrayStringer[*T](stringer)(vv)
		}
		var sb strings.Builder
		sb.WriteString("[")
		if len(vv) > 0 {
			sb.WriteString(stringer(vv[0]))
			for _, x := range vv[1:] {
				sb.WriteString(", ")
				sb.WriteString(stringer(x))
			}
		}
		sb.WriteString("]")
		return sb.String()
	}
}

func javaClassStringer(v any) string {
	return fmt.Sprintf("Java Class: %v", v)
}

func dateTimeStringer(v any) string {
	switch vv := v.(type) {
	case time.Time:
		return vv.Format(time.RFC3339)
	case *time.Time:
		if vv == nil {
			return ValueNil
		}
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
	return "?datetime?"
}

func portableStringer(v any) string {
	vv := v.(*GenericPortable)
	return vv.Text()
}

var ValueToText map[int32]Stringer

func init() {
	ValueToText = map[int32]Stringer{
		TypeNil:        staticStringer(ValueNil),
		TypeSkip:       staticStringer(ValueSkip),
		TypeNotDecoded: staticStringer(ValueNotDecoded),

		TypePortable: portableStringer,
		// TypeDataSerializable

		TypeByte:         ptrStringer[byte](), // +
		TypeBool:         ptrStringer[bool](), // +
		TypeUInt16:       ptrStringer[uint16](),
		TypeInt16:        ptrStringer[int16](),
		TypeInt32:        ptrStringer[int32](),
		TypeInt64:        ptrStringer[int64](),
		TypeFloat32:      ptrStringer[float32](),
		TypeFloat64:      ptrStringer[float64](),
		TypeString:       ptrStringer[string](), // +
		TypeByteArray:    arrayStringer[uint8](sprintStringer),
		TypeBoolArray:    arrayStringer[bool](sprintStringer),
		TypeUInt16Array:  arrayStringer[uint16](sprintStringer),
		TypeInt16Array:   arrayStringer[int16](sprintStringer),
		TypeInt32Array:   arrayStringer[int32](sprintStringer),
		TypeInt64Array:   arrayStringer[int64](sprintStringer),
		TypeFloat32Array: arrayStringer[float32](sprintStringer),
		TypeFloat64Array: arrayStringer[float64](sprintStringer),
		TypeStringArray:  arrayStringer[string](sprintStringer),
		TypeUUID:         sprintStringer,

		// TypeSimpleEntry
		// TypeSimpleImmutableEntry

		TypeJavaClass:          javaClassStringer,
		TypeJavaDate:           dateTimeStringer,
		TypeJavaBigInteger:     sprintStringer,
		TypeJavaDecimal:        ptrStringer[types.Decimal](),          // +
		TypeJavaArray:          arrayAnyStringer[any](sprintStringer), // +
		TypeJavaArrayList:      arrayAnyStringer[any](sprintStringer), // +
		TypeJavaLinkedList:     arrayAnyStringer[any](sprintStringer), // +
		TypeJavaLocalDate:      dateTimeStringer,
		TypeJavaLocalTime:      dateTimeStringer,
		TypeJavaLocalDateTime:  dateTimeStringer,
		TypeJavaOffsetDateTime: dateTimeStringer, // +

		TypeJSONSerialization: sprintStringer,

		//
		TypeDecimalArray:            arrayAnyStringer[types.Decimal](ptrStringer[types.Decimal]()),
		TypeJavaLocalTimeArray:      arrayAnyStringer[types.LocalTime](dateTimeStringer),
		TypeJavaLocalDateArray:      arrayAnyStringer[types.LocalDate](dateTimeStringer),
		TypeJavaLocalDateTimeArray:  arrayAnyStringer[types.LocalDateTime](dateTimeStringer),
		TypeJavaOffsetDateTimeArray: arrayAnyStringer[types.OffsetDateTime](dateTimeStringer),
		TypeCompactArray:            arrayAnyStringer[any](sprintStringer),
		TypeInt8:                    ptrStringer[int8](),
		TypeInt8Array:               arrayStringer[int8](sprintStringer),
	}
}
