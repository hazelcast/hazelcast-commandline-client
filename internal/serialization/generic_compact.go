package serialization

import (
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type compactFieldReader func(r serialization.CompactReader, field string) any

type compactFieldWriter func(w serialization.CompactWriter, field string, value any)

type GenericCompact struct {
	ValueType     reflect.Type
	ValueTypeName string
	Fields        []CompactField
	readers       []compactFieldReader
	writers       []compactFieldWriter
}

func NewGenericCompact(value GenericCompact) (*GenericCompact, error) {
	rs := make([]compactFieldReader, len(value.Fields))
	ws := make([]compactFieldWriter, len(value.Fields))
	for i, f := range value.Fields {
		if f.Type < CompactFieldType(serialization.FieldKindNotAvailable) || f.Type > CompactFieldType(serialization.FieldKindArrayOfNullableFloat64) {
			return nil, fmt.Errorf("invalid portable type: %d", f.Type)
		}
		r, ok := compactReaders[serialization.FieldKind(f.Type)]
		if !ok {
			return nil, fmt.Errorf("reader not found for compact type: %d", f.Type)
		}
		rs[i] = r
		// writing is disabled for now --YT
	}
	return &GenericCompact{
		Fields:        value.Fields,
		ValueType:     value.makeType(),
		ValueTypeName: value.ValueTypeName,
		readers:       rs,
		writers:       ws,
	}, nil

}

func (cm GenericCompact) makeType() reflect.Type {
	fs := make([]reflect.StructField, len(cm.Fields))
	for i, f := range cm.Fields {
		fs[i] = reflect.StructField{
			Name: fmt.Sprintf("Field%03d", i),
			Type: fieldKindToType[serialization.FieldKind(f.Type)],
			Tag:  reflect.StructTag(fmt.Sprintf("json:\"%s\"", f.Name)),
		}
	}
	return reflect.StructOf(fs)
}

func (cm GenericCompact) Type() reflect.Type {
	return cm.ValueType
}

func (cm GenericCompact) TypeName() string {
	return cm.ValueTypeName
}

func (cm GenericCompact) Read(reader serialization.CompactReader) interface{} {
	rs := cm.readers
	v := reflect.New(cm.ValueType)
	for i, f := range cm.Fields {
		value := reflect.ValueOf(rs[i](reader, f.Name))
		if value.Interface() == nil {
			continue
		}
		v.Elem().Field(i).Set(value)
	}
	return v.Interface()
}

func (cm GenericCompact) Write(writer serialization.CompactWriter, value interface{}) {
	// TODO: implement me when compact write is supported
	panic("implement me")
}

var fieldKindToType map[serialization.FieldKind]reflect.Type

func init() {
	var a any
	var b bool
	var i8 int8
	var i16 int16
	var i32 int32
	var i64 int64
	var f32 float32
	var f64 float64
	var t time.Time
	fieldKindToType = map[serialization.FieldKind]reflect.Type{
		serialization.FieldKindNotAvailable:                 nil,
		serialization.FieldKindBoolean:                      reflect.TypeOf(b),
		serialization.FieldKindArrayOfBoolean:               reflect.TypeOf([]bool{}),
		serialization.FieldKindInt8:                         reflect.TypeOf(i8),
		serialization.FieldKindArrayOfInt8:                  reflect.TypeOf([]int8{}),
		serialization.FieldKindInt16:                        reflect.TypeOf(i16),
		serialization.FieldKindArrayOfInt16:                 reflect.TypeOf([]int16{}),
		serialization.FieldKindInt32:                        reflect.TypeOf(i32),
		serialization.FieldKindArrayOfInt32:                 reflect.TypeOf([]int32{}),
		serialization.FieldKindInt64:                        reflect.TypeOf(i64),
		serialization.FieldKindArrayOfInt64:                 reflect.TypeOf([]int64{}),
		serialization.FieldKindFloat32:                      reflect.TypeOf(f32),
		serialization.FieldKindArrayOfFloat32:               reflect.TypeOf([]float32{}),
		serialization.FieldKindFloat64:                      reflect.TypeOf(f64),
		serialization.FieldKindArrayOfFloat64:               reflect.TypeOf([]float64{}),
		serialization.FieldKindString:                       reflect.TypeOf(""),
		serialization.FieldKindArrayOfString:                reflect.TypeOf([]string{}),
		serialization.FieldKindDecimal:                      reflect.TypeOf(types.NewDecimal(new(big.Int), 0)),
		serialization.FieldKindArrayOfDecimal:               reflect.TypeOf([]types.Decimal{}),
		serialization.FieldKindTime:                         reflect.TypeOf(types.LocalTime(t)),
		serialization.FieldKindArrayOfTime:                  reflect.TypeOf([]types.LocalTime{}),
		serialization.FieldKindDate:                         reflect.TypeOf(types.LocalTime(t)),
		serialization.FieldKindArrayOfDate:                  reflect.TypeOf([]types.LocalDate{}),
		serialization.FieldKindTimestamp:                    reflect.TypeOf(types.LocalTime(t)),
		serialization.FieldKindArrayOfTimestamp:             reflect.TypeOf([]types.LocalDateTime{}),
		serialization.FieldKindTimestampWithTimezone:        reflect.TypeOf(types.LocalTime(t)),
		serialization.FieldKindArrayOfTimestampWithTimezone: reflect.TypeOf([]types.OffsetDateTime{}),
		serialization.FieldKindCompact:                      reflect.TypeOf(a),
		serialization.FieldKindArrayOfCompact:               reflect.TypeOf([]any{}),
		serialization.FieldKindNullableBoolean:              reflect.TypeOf(&b),
		serialization.FieldKindArrayOfNullableBoolean:       reflect.TypeOf([]*bool{}),
		serialization.FieldKindNullableInt8:                 reflect.TypeOf(&i8),
		serialization.FieldKindArrayOfNullableInt8:          reflect.TypeOf([]*int8{}),
		serialization.FieldKindNullableInt16:                reflect.TypeOf(&i16),
		serialization.FieldKindArrayOfNullableInt16:         reflect.TypeOf([]*int16{}),
		serialization.FieldKindNullableInt32:                reflect.TypeOf(&i32),
		serialization.FieldKindArrayOfNullableInt32:         reflect.TypeOf([]*int32{}),
		serialization.FieldKindNullableInt64:                reflect.TypeOf(&i64),
		serialization.FieldKindArrayOfNullableInt64:         reflect.TypeOf([]*int64{}),
		serialization.FieldKindNullableFloat32:              reflect.TypeOf(&f32),
		serialization.FieldKindArrayOfNullableFloat32:       reflect.TypeOf([]*float32{}),
		serialization.FieldKindNullableFloat64:              reflect.TypeOf(&f64),
		serialization.FieldKindArrayOfNullableFloat64:       reflect.TypeOf([]*float64{}),
	}
}
