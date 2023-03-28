package serialization

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type compactFieldReader func(r serialization.CompactReader, field string) any

type SchemaInfo struct {
	Type     reflect.Type
	TypeName string
}

type GenericCompactDeserializer struct {
	SchemaInfos *sync.Map
}

func NewGenericCompactDeserializer() *GenericCompactDeserializer {
	return &GenericCompactDeserializer{
		SchemaInfos: &sync.Map{},
	}
}

func (cm *GenericCompactDeserializer) Read(schema *hazelcast.Schema, reader serialization.CompactReader) interface{} {
	v := reflect.New(cm.makeType(schema))
	for i, fd := range schema.FieldDefinitions() {
		r := compactReaders[fd.Kind]
		value := reflect.ValueOf(r(reader, fd.Name))
		if value.Interface() == nil {
			continue
		}
		v.Elem().Field(i).Set(value)
	}
	return v.Interface()
}

func (cm *GenericCompactDeserializer) makeType(schema *hazelcast.Schema) reflect.Type {
	var t reflect.Type
	vi, ok := cm.SchemaInfos.Load(schema.ID())
	if ok {
		v := vi.(SchemaInfo)
		t = v.Type
	} else {
		fds := schema.FieldDefinitions()
		fs := make([]reflect.StructField, len(fds))
		for i, f := range fds {
			fs[i] = reflect.StructField{
				Name: fmt.Sprintf("Field%03d", i),
				Type: fieldKindToType[f.Kind],
				Tag:  reflect.StructTag(fmt.Sprintf("json:\"%s\"", f.Name)),
			}
		}
		t = reflect.StructOf(fs)
		cm.SchemaInfos.Store(schema.ID(), SchemaInfo{
			Type:     t,
			TypeName: schema.TypeName,
		})
	}
	return t
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
		serialization.FieldKindDate:                         reflect.TypeOf(types.LocalDate(t)),
		serialization.FieldKindArrayOfDate:                  reflect.TypeOf([]types.LocalDate{}),
		serialization.FieldKindTimestamp:                    reflect.TypeOf(types.LocalDateTime(t)),
		serialization.FieldKindArrayOfTimestamp:             reflect.TypeOf([]types.LocalDateTime{}),
		serialization.FieldKindTimestampWithTimezone:        reflect.TypeOf(types.OffsetDateTime(t)),
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
