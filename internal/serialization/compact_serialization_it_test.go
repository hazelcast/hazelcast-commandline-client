package serialization_test

import (
	"context"
	"math"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/map"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestCompactSerialization(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "CompactOtherArrays", f: compactOtherArraysTest},
		{name: "CompactOthers", f: compactOthersTest},
		{name: "CompactPrimitiveArrays", f: compactPrimitiveArraysTest},
		{name: "CompactPrimitives", f: compactPrimitivesTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func compactPrimitiveArraysTest(t *testing.T) {
	tcx := it.TestContext{
		T: t,
		ConfigCallback: func(tcx it.TestContext) {
			tcx.ClientConfig.Serialization.Compact.SetSerializers(primitiveArraysSerializer{})
		},
	}
	tcx.Tester(func(tcx it.TestContext) {
		b := true
		i8 := int8(8)
		value := primitiveArrays{
			fullBoolArray:         []bool{true, false},
			fullInt8Array:         []int8{math.MinInt8, 0, math.MaxInt8},
			fullInt16Array:        []int16{math.MinInt16, 0, math.MaxInt16},
			fullInt32Array:        []int32{math.MinInt32, 0, math.MaxInt32},
			fullInt64Array:        []int64{math.MinInt64, 0, math.MinInt64},
			fullFloat32Array:      []float32{math.SmallestNonzeroFloat32, 0, math.MaxFloat32},
			fullFloat64Array:      []float64{math.SmallestNonzeroFloat64, 0, math.MaxFloat64},
			fullNullableBoolArray: []*bool{&b, nil},
			fullNullableInt8Array: []*int8{nil, &i8},
		}
		ctx := context.Background()
		it.WithMap(tcx, func(m *hazelcast.Map) {
			check.Must(m.Set(ctx, "value", value))
			testCases := []struct {
				format string
				target string
			}{
				{
					format: "delimited",
					target: "{emptyBoolArray:[]; emptyFloat32Array:[]; emptyFloat64Array:[]; emptyInt16Array:[]; emptyInt32Array:[]; emptyInt64Array:[]; emptyInt8Array:[]; emptyNullableBoolArray:[]; emptyNullableInt8Array:[]; fullBoolArray:[true, false]; fullFloat32Array:[1e-45, 0, 3.4028235e+38]; fullFloat64Array:[5e-324, 0, 1.7976931348623157e+308]; fullInt16Array:[-32768, 0, 32767]; fullInt32Array:[-2147483648, 0, 2147483647]; fullInt64Array:[-9223372036854775808, 0, -9223372036854775808]; fullInt8Array:[-128, 0, 127]; fullNullableBoolArray:[true, -]; fullNullableInt8Array:[-, 8]}\n",
				},
				{
					format: "json",
					target: `{"this":{"emptyBoolArray":[],"emptyFloat32Array":[],"emptyFloat64Array":[],"emptyInt16Array":[],"emptyInt32Array":[],"emptyInt64Array":[],"emptyInt8Array":[],"emptyNullableBoolArray":[],"emptyNullableInt8Array":[],"fullBoolArray":[true,false],"fullFloat32Array":[1e-45,0,3.4028235e+38],"fullFloat64Array":[5e-324,0,1.7976931348623157e+308],"fullInt16Array":[-32768,0,32767],"fullInt32Array":[-2147483648,0,2147483647],"fullInt64Array":[-9223372036854775808,0,-9223372036854775808],"fullInt8Array":[-128,0,127],"fullNullableBoolArray":[true,null],"fullNullableInt8Array":[null,8]}}` + "\n",
				},
				{
					format: "csv",
					target: "this\n" + `"{emptyBoolArray:[]; emptyFloat32Array:[]; emptyFloat64Array:[]; emptyInt16Array:[]; emptyInt32Array:[]; emptyInt64Array:[]; emptyInt8Array:[]; emptyNullableBoolArray:[]; emptyNullableInt8Array:[]; fullBoolArray:[true, false]; fullFloat32Array:[1e-45, 0, 3.4028235e+38]; fullFloat64Array:[5e-324, 0, 1.7976931348623157e+308]; fullInt16Array:[-32768, 0, 32767]; fullInt32Array:[-2147483648, 0, 2147483647]; fullInt64Array:[-9223372036854775808, 0, -9223372036854775808]; fullInt8Array:[-128, 0, 127]; fullNullableBoolArray:[true, -]; fullNullableInt8Array:[-, 8]}"` + "\n",
				},

				{
					format: "table",
					target: "testdata/compact_primitive_arrays_table.txt",
				},
			}
			for _, tc := range testCases {
				tc := tc
				t.Run(tc.format, func(t *testing.T) {
					tcx.T = t
					tcx.WithReset(func() {
						ctx := context.Background()
						check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "get", "value", "-q", "-f", tc.format))
						if tc.format == "table" {
							tcx.AssertStdoutDollarWithPath(tc.target)
						} else {
							tcx.AssertStdoutEquals(tc.target)
						}
					})
				})
			}
		})

	})
}

func compactPrimitivesTest(t *testing.T) {
	tcx := it.TestContext{
		T: t,
		ConfigCallback: func(tcx it.TestContext) {
			tcx.ClientConfig.Serialization.Compact.SetSerializers(primitivesSerializer{})
		},
	}
	tcx.Tester(func(tcx it.TestContext) {
		i8 := int8(8)
		b := false
		value := primitives{
			valueInt8:       i8,
			nullInt8NotNull: &i8,
			valueBool:       true,
			nullBoolNotNull: &b,
		}
		ctx := context.Background()
		it.WithMap(tcx, func(m *hazelcast.Map) {
			check.Must(m.Set(ctx, "value", value))
			testCases := []struct {
				format string
				target string
			}{
				{
					format: "delimited",
					target: "{nullBoolNotNull:false; nullBoolNull:-; nullInt8NotNull:8; nullInt8Null:-; valueBool:true; valueInt8:8}\n",
				},
				{
					format: "json",
					target: `{"this":{"nullBoolNotNull":false,"nullBoolNull":null,"nullInt8NotNull":8,"nullInt8Null":null,"valueBool":true,"valueInt8":8}}` + "\n",
				},
				{
					format: "csv",
					target: "this\n{nullBoolNotNull:false; nullBoolNull:-; nullInt8NotNull:8; nullInt8Null:-; valueBool:true; valueInt8:8}\n",
				},
				{
					format: "table",
					target: "testdata/compact_primitives_table.txt",
				},
			}
			for _, tc := range testCases {
				tc := tc
				t.Run(tc.format, func(t *testing.T) {
					tcx.T = t
					tcx.WithReset(func() {
						ctx := context.Background()
						check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "get", "value", "-q", "-f", tc.format))
						if tc.format == "table" {
							tcx.AssertStdoutDollarWithPath(tc.target)
						} else {
							tcx.AssertStdoutEquals(tc.target)
						}
					})
				})
			}
		})

	})
}

func compactOthersTest(t *testing.T) {
	tcx := it.TestContext{
		T: t,
		ConfigCallback: func(tcx it.TestContext) {
			tcx.ClientConfig.Serialization.Compact.SetSerializers(othersSerializer{})
		},
	}
	tcx.Tester(func(tcx it.TestContext) {
		s := "foobar"
		dtz := time.Date(2023, 4, 5, 12, 33, 45, 46, time.UTC)
		dc := types.NewDecimal(big.NewInt(1234), 56)
		value := others{
			nullStringNotNull:     &s,
			offsetDateTimeNotNull: (*types.OffsetDateTime)(&dtz),
			decimalNotNull:        &dc,
		}
		ctx := context.Background()
		it.WithMap(tcx, func(m *hazelcast.Map) {
			check.Must(m.Set(ctx, "value", value))
			testCases := []struct {
				format string
				target string
			}{
				{
					format: "delimited",
					target: "{decimalNotNull:1.234E-53; decimalNull:-; nullStringNotNull:foobar; nullStringNull:-; offsetDateTimeNotNull:2023-04-05T12:33:45Z; offsetDateTimeNull:-}\n",
				},
				{
					format: "json",
					target: `{"this":{"decimalNotNull":"1.234E-53","decimalNull":null,"nullStringNotNull":"foobar","nullStringNull":null,"offsetDateTimeNotNull":"2023-04-05T12:33:45Z","offsetDateTimeNull":null}}` + "\n",
				},
				{
					format: "csv",
					target: "this\n" + `{decimalNotNull:1.234E-53; decimalNull:-; nullStringNotNull:foobar; nullStringNull:-; offsetDateTimeNotNull:2023-04-05T12:33:45Z; offsetDateTimeNull:-}` + "\n",
				},
				{
					format: "table",
					target: "testdata/compact_others_table.txt",
				},
			}
			for _, tc := range testCases {
				tc := tc
				t.Run(tc.format, func(t *testing.T) {
					tcx.T = t
					tcx.WithReset(func() {
						ctx := context.Background()
						check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "get", "value", "-q", "-f", tc.format))
						if tc.format == "table" {
							tcx.AssertStdoutDollarWithPath(tc.target)
						} else {
							tcx.AssertStdoutEquals(tc.target)
						}
					})
				})
			}
		})

	})
}

func compactOtherArraysTest(t *testing.T) {
	dt1 := time.Date(2023, 1, 2, 3, 4, 5, 6, time.UTC)
	dt2 := time.Date(2022, 2, 3, 4, 5, 6, 7, time.UTC)
	value := otherArrays{
		fullTimeArray: []*types.LocalTime{
			(*types.LocalTime)(&dt1),
			(*types.LocalTime)(&dt2),
		},
		fullDateArray: []*types.LocalDate{
			(*types.LocalDate)(&dt1),
			(*types.LocalDate)(&dt2),
		},
		fullTimestampTimeArray: []*types.LocalDateTime{
			(*types.LocalDateTime)(&dt1),
			(*types.LocalDateTime)(&dt2),
		},
		fullTimestampWithTimezoneArray: []*types.OffsetDateTime{
			(*types.OffsetDateTime)(&dt1),
			(*types.OffsetDateTime)(&dt2),
		},
		fullDecimalArray: []*types.Decimal{
			ptr(types.NewDecimal(big.NewInt(1234), 67)),
			ptr(types.NewDecimal(big.NewInt(4567), 89)),
		},
		fullCompactArray: []any{
			simpleObj{value: ptr("obj1")},
			simpleObj{value: ptr("obj2")},
		},
	}
	testCases := []struct {
		format string
		target string
	}{
		{
			format: "delimited",
			target: `{emptyCompactArray:[]; emptyDateArray:[]; emptyDecimalArray:[]; emptyTimeArray:[]; emptyTimestampArray:[]; emptyTimestampWithTimezoneArray:[]; fullCompactArray:[{value:obj1}, {value:obj2}]; fullDateArray:[2023-01-02, 2022-02-03]; fullDecimalArray:[1.234E-64, 4.567E-86]; fullTimeArray:[03:04:05, 04:05:06]; fullTimestampTimeArray:[2023-01-02 03:04:05, 2022-02-03 04:05:06]; fullTimestampWithTimezoneArray:[2023-01-02T03:04:05Z, 2022-02-03T04:05:06Z]}` + "\n",
		},
		{
			format: "json",
			target: `{"this":{"emptyCompactArray":[],"emptyDateArray":[],"emptyDecimalArray":[],"emptyTimeArray":[],"emptyTimestampArray":[],"emptyTimestampWithTimezoneArray":[],"fullCompactArray":[{"value":"obj1"},{"value":"obj2"}],"fullDateArray":["2023-01-02","2022-02-03"],"fullDecimalArray":["1.234E-64","4.567E-86"],"fullTimeArray":["03:04:05","04:05:06"],"fullTimestampTimeArray":["2023-01-02 03:04:05","2022-02-03 04:05:06"],"fullTimestampWithTimezoneArray":["2023-01-02T03:04:05Z","2022-02-03T04:05:06Z"]}}` + "\n",
		},
		{
			format: "csv",
			target: "this\n" + `"{emptyCompactArray:[]; emptyDateArray:[]; emptyDecimalArray:[]; emptyTimeArray:[]; emptyTimestampArray:[]; emptyTimestampWithTimezoneArray:[]; fullCompactArray:[{value:obj1}, {value:obj2}]; fullDateArray:[2023-01-02, 2022-02-03]; fullDecimalArray:[1.234E-64, 4.567E-86]; fullTimeArray:[03:04:05, 04:05:06]; fullTimestampTimeArray:[2023-01-02 03:04:05, 2022-02-03 04:05:06]; fullTimestampWithTimezoneArray:[2023-01-02T03:04:05Z, 2022-02-03T04:05:06Z]}"` + "\n",
		},
		{
			format: "table",
			target: "testdata/compact_other_arrays_table.txt",
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.format, func(t *testing.T) {
			tcx := it.TestContext{
				T: t,
				ConfigCallback: func(tcx it.TestContext) {
					tcx.ClientConfig.Serialization.Compact.SetSerializers(
						otherArraysSerializer{},
						simpleObjSerializer{},
					)
				},
			}
			tcx.Tester(func(tcx it.TestContext) {
				it.WithMap(tcx, func(m *hazelcast.Map) {
					check.Must(m.Set(ctx, "value", value))
					tcx.WithReset(func() {
						ctx := context.Background()
						check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "get", "value", "-q", "-f", tc.format))
						if tc.format == "table" {
							tcx.AssertStdoutDollarWithPath(tc.target)
						} else {
							tcx.AssertStdoutEquals(tc.target)
						}
					})
				})
			})
		})
	}
}

type primitives struct {
	valueInt8       int8
	nullInt8Null    *int8
	nullInt8NotNull *int8
	valueBool       bool
	nullBoolNull    *bool
	nullBoolNotNull *bool
}

type primitivesSerializer struct{}

func (s primitivesSerializer) Type() reflect.Type {
	return reflect.TypeOf(primitives{})
}

func (s primitivesSerializer) TypeName() string {
	return "primitives"
}

func (s primitivesSerializer) Read(r serialization.CompactReader) interface{} {
	return primitives{
		valueInt8:       r.ReadInt8("valueInt8"),
		nullInt8Null:    r.ReadNullableInt8("nullInt8Null"),
		nullInt8NotNull: r.ReadNullableInt8("nullInt8NotNull"),
		valueBool:       r.ReadBoolean("valueBool"),
		nullBoolNull:    r.ReadNullableBoolean("nullBoolNull"),
		nullBoolNotNull: r.ReadNullableBoolean("nullBoolNotNull"),
	}
}

func (s primitivesSerializer) Write(w serialization.CompactWriter, v interface{}) {
	vv := v.(primitives)
	w.WriteInt8("valueInt8", vv.valueInt8)
	w.WriteNullableInt8("nullInt8Null", vv.nullInt8Null)
	w.WriteNullableInt8("nullInt8NotNull", vv.nullInt8NotNull)
	w.WriteBoolean("valueBool", vv.valueBool)
	w.WriteNullableBoolean("nullBoolNull", vv.nullBoolNull)
	w.WriteNullableBoolean("nullBoolNotNull", vv.nullBoolNotNull)
}

type primitiveArrays struct {
	emptyBoolArray         []bool
	fullBoolArray          []bool
	emptyInt8Array         []int8
	fullInt8Array          []int8
	emptyInt16Array        []int16
	fullInt16Array         []int16
	emptyInt32Array        []int32
	fullInt32Array         []int32
	emptyInt64Array        []int64
	fullInt64Array         []int64
	emptyFloat32Array      []float32
	fullFloat32Array       []float32
	emptyFloat64Array      []float64
	fullFloat64Array       []float64
	emptyNullableBoolArray []*bool
	fullNullableBoolArray  []*bool
	emptyNullableInt8Array []*int8
	fullNullableInt8Array  []*int8
}

type primitiveArraysSerializer struct{}

func (s primitiveArraysSerializer) Type() reflect.Type {
	return reflect.TypeOf(primitiveArrays{})
}

func (s primitiveArraysSerializer) TypeName() string {
	return "primitiveArrays"
}

func (s primitiveArraysSerializer) Read(r serialization.CompactReader) interface{} {
	return primitiveArrays{
		emptyBoolArray:         r.ReadArrayOfBoolean("emptyBoolArray"),
		fullBoolArray:          r.ReadArrayOfBoolean("fullBoolArray"),
		emptyInt8Array:         r.ReadArrayOfInt8("emptyInt8Array"),
		fullInt8Array:          r.ReadArrayOfInt8("fullInt8Array"),
		emptyInt16Array:        r.ReadArrayOfInt16("emptyInt16Array"),
		fullInt16Array:         r.ReadArrayOfInt16("fullInt16Array"),
		emptyInt32Array:        r.ReadArrayOfInt32("emptyInt32Array"),
		fullInt32Array:         r.ReadArrayOfInt32("fullInt32Array"),
		emptyInt64Array:        r.ReadArrayOfInt64("emptyInt64Array"),
		fullInt64Array:         r.ReadArrayOfInt64("fullInt64Array"),
		emptyFloat32Array:      r.ReadArrayOfFloat32("emptyFloat32Array"),
		fullFloat32Array:       r.ReadArrayOfFloat32("fullFloat32Array"),
		emptyFloat64Array:      r.ReadArrayOfFloat64("emptyFloat64Array"),
		fullFloat64Array:       r.ReadArrayOfFloat64("fullFloat64Array"),
		emptyNullableBoolArray: r.ReadArrayOfNullableBoolean("emptyNullableBoolArray"),
		fullNullableBoolArray:  r.ReadArrayOfNullableBoolean("fullNullableBoolArray"),
		emptyNullableInt8Array: r.ReadArrayOfNullableInt8("emptyNullableInt8Array"),
		fullNullableInt8Array:  r.ReadArrayOfNullableInt8("fullNullableInt8Array"),
	}
}

func (s primitiveArraysSerializer) Write(w serialization.CompactWriter, v interface{}) {
	vv := v.(primitiveArrays)
	w.WriteArrayOfBoolean("emptyBoolArray", vv.emptyBoolArray)
	w.WriteArrayOfBoolean("fullBoolArray", vv.fullBoolArray)
	w.WriteArrayOfInt8("emptyInt8Array", vv.emptyInt8Array)
	w.WriteArrayOfInt8("fullInt8Array", vv.fullInt8Array)
	w.WriteArrayOfInt16("emptyInt16Array", vv.emptyInt16Array)
	w.WriteArrayOfInt16("fullInt16Array", vv.fullInt16Array)
	w.WriteArrayOfInt32("emptyInt32Array", vv.emptyInt32Array)
	w.WriteArrayOfInt32("fullInt32Array", vv.fullInt32Array)
	w.WriteArrayOfInt64("emptyInt64Array", vv.emptyInt64Array)
	w.WriteArrayOfInt64("fullInt64Array", vv.fullInt64Array)
	w.WriteArrayOfFloat32("emptyFloat32Array", vv.emptyFloat32Array)
	w.WriteArrayOfFloat32("fullFloat32Array", vv.fullFloat32Array)
	w.WriteArrayOfFloat64("emptyFloat64Array", vv.emptyFloat64Array)
	w.WriteArrayOfFloat64("fullFloat64Array", vv.fullFloat64Array)
	w.WriteArrayOfNullableBoolean("fullNullableBoolArray", vv.fullNullableBoolArray)
	w.WriteArrayOfNullableBoolean("emptyNullableBoolArray", vv.emptyNullableBoolArray)
	w.WriteArrayOfNullableInt8("emptyNullableInt8Array", vv.emptyNullableInt8Array)
	w.WriteArrayOfNullableInt8("fullNullableInt8Array", vv.fullNullableInt8Array)
}

type others struct {
	nullStringNull        *string
	nullStringNotNull     *string
	offsetDateTimeNull    *types.OffsetDateTime
	offsetDateTimeNotNull *types.OffsetDateTime
	decimalNull           *types.Decimal
	decimalNotNull        *types.Decimal
}

type othersSerializer struct{}

func (s othersSerializer) Type() reflect.Type {
	return reflect.TypeOf(others{})
}

func (s othersSerializer) TypeName() string {
	return "others"
}

func (s othersSerializer) Read(r serialization.CompactReader) interface{} {
	return others{
		nullStringNull:        r.ReadString("nullStringNull"),
		nullStringNotNull:     r.ReadString("nullStringNotNull"),
		offsetDateTimeNull:    r.ReadTimestampWithTimezone("offsetDateTimeNull"),
		offsetDateTimeNotNull: r.ReadTimestampWithTimezone("offsetDateTimeNotNull"),
		decimalNull:           r.ReadDecimal("decimalNull"),
		decimalNotNull:        r.ReadDecimal("decimalNotNull"),
	}
}

func (s othersSerializer) Write(w serialization.CompactWriter, v interface{}) {
	vv := v.(others)
	w.WriteString("nullStringNull", vv.nullStringNull)
	w.WriteString("nullStringNotNull", vv.nullStringNotNull)
	w.WriteTimestampWithTimezone("offsetDateTimeNull", vv.offsetDateTimeNull)
	w.WriteTimestampWithTimezone("offsetDateTimeNotNull", vv.offsetDateTimeNotNull)
	w.WriteDecimal("decimalNull", vv.decimalNull)
	w.WriteDecimal("decimalNotNull", vv.decimalNotNull)
}

type simpleObj struct {
	value *string
}

type simpleObjSerializer struct{}

func (s simpleObjSerializer) Type() reflect.Type {
	return reflect.TypeOf(simpleObj{})
}

func (s simpleObjSerializer) TypeName() string {
	return "simpleObj"
}

func (s simpleObjSerializer) Read(r serialization.CompactReader) interface{} {
	return simpleObj{
		value: r.ReadString("value"),
	}
}

func (s simpleObjSerializer) Write(w serialization.CompactWriter, v interface{}) {
	vv := v.(simpleObj)
	w.WriteString("value", vv.value)
}

type otherArrays struct {
	emptyTimeArray                  []*types.LocalTime
	fullTimeArray                   []*types.LocalTime
	emptyDateArray                  []*types.LocalDate
	fullDateArray                   []*types.LocalDate
	emptyTimestampArray             []*types.LocalDateTime
	fullTimestampTimeArray          []*types.LocalDateTime
	emptyTimestampWithTimezoneArray []*types.OffsetDateTime
	fullTimestampWithTimezoneArray  []*types.OffsetDateTime
	emptyDecimalArray               []*types.Decimal
	fullDecimalArray                []*types.Decimal
	emptyCompactArray               []any
	fullCompactArray                []any
}

type otherArraysSerializer struct{}

func (s otherArraysSerializer) Type() reflect.Type {
	return reflect.TypeOf(otherArrays{})
}

func (s otherArraysSerializer) TypeName() string {
	return "otherArrays"
}

func (s otherArraysSerializer) Read(r serialization.CompactReader) interface{} {
	return otherArrays{
		emptyTimeArray:                  r.ReadArrayOfTime("emptyTimeArray"),
		fullTimeArray:                   r.ReadArrayOfTime("fullTimeArray"),
		emptyDateArray:                  r.ReadArrayOfDate("emptyDateArray"),
		fullDateArray:                   r.ReadArrayOfDate("fullDateArray"),
		emptyTimestampArray:             r.ReadArrayOfTimestamp("emptyTimestampArray"),
		fullTimestampTimeArray:          r.ReadArrayOfTimestamp("fullTimestampTimeArray"),
		emptyTimestampWithTimezoneArray: r.ReadArrayOfTimestampWithTimezone("emptyTimestampWithTimezoneArray"),
		fullTimestampWithTimezoneArray:  r.ReadArrayOfTimestampWithTimezone("fullTimestampWithTimezoneArray"),
		emptyDecimalArray:               r.ReadArrayOfDecimal("emptyDecimalArray"),
		fullDecimalArray:                r.ReadArrayOfDecimal("fullDecimalArray"),
		emptyCompactArray:               r.ReadArrayOfCompact("emptyCompactArray"),
		fullCompactArray:                r.ReadArrayOfCompact("fullCompactArray"),
	}
}

func (s otherArraysSerializer) Write(w serialization.CompactWriter, v interface{}) {
	vv := v.(otherArrays)
	w.WriteArrayOfTime("emptyTimeArray", vv.emptyTimeArray)
	w.WriteArrayOfTime("fullTimeArray", vv.fullTimeArray)
	w.WriteArrayOfDate("emptyDateArray", vv.emptyDateArray)
	w.WriteArrayOfDate("fullDateArray", vv.fullDateArray)
	w.WriteArrayOfTimestamp("emptyTimestampArray", vv.emptyTimestampArray)
	w.WriteArrayOfTimestamp("fullTimestampTimeArray", vv.fullTimestampTimeArray)
	w.WriteArrayOfTimestampWithTimezone("emptyTimestampWithTimezoneArray", vv.emptyTimestampWithTimezoneArray)
	w.WriteArrayOfTimestampWithTimezone("fullTimestampWithTimezoneArray", vv.fullTimestampWithTimezoneArray)
	w.WriteArrayOfDecimal("emptyDecimalArray", vv.emptyDecimalArray)
	w.WriteArrayOfDecimal("fullDecimalArray", vv.fullDecimalArray)
	w.WriteArrayOfCompact("emptyCompactArray", vv.emptyCompactArray)
	w.WriteArrayOfCompact("fullCompactArray", vv.fullCompactArray)
}

func ptr[T any](v T) *T {
	return &v
}
