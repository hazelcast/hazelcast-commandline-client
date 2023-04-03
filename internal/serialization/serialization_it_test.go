package serialization_test

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"

	_ "github.com/hazelcast/hazelcast-commandline-client/base/commands/map"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestCompactSerialization(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "PrimitiveArrays", f: primitiveArraysTest},
		{name: "Primitives", f: primitivesTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func primitiveArraysTest(t *testing.T) {
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
			fullNullableBoolArray: []*bool{&b},
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
					target: "emptyBoolArray:[]; emptyFloat32Array:[]; emptyFloat64Array:[]; emptyInt16Array:[]; emptyInt32Array:[]; emptyInt64Array:[]; emptyInt8Array:[]; emptyNullableBoolArray:[]; emptyNullableInt8Array:[]; fullBoolArray:[true, false]; fullFloat32Array:[]; fullFloat64Array:[]; fullInt16Array:[-32768, 0, 32767]; fullInt32Array:[-2147483648, 0, 2147483647]; fullInt64Array:[-9223372036854775808, 0, -9223372036854775808]; fullInt8Array:[-128, 0, 127]; fullNullableBoolArray:[true]; fullNullableInt8Array:[-, 8]\n",
				},
				{
					format: "json",
					target: `{"this":{"emptyBoolArray":null,"emptyFloat32Array":null,"emptyFloat64Array":null,"emptyInt16Array":null,"emptyInt32Array":null,"emptyInt64Array":null,"emptyInt8Array":null,"emptyNullableBoolArray":null,"emptyNullableInt8Array":null,"fullBoolArray":[true,false],"fullFloat32Array":null,"fullFloat64Array":null,"fullInt16Array":[-32768,0,32767],"fullInt32Array":[-2147483648,0,2147483647],"fullInt64Array":[-9223372036854775808,0,-9223372036854775808],"fullInt8Array":[-128,0,127],"fullNullableBoolArray":[true],"fullNullableInt8Array":[null,8]}}` + "\n",
				},
				{
					format: "csv",
					target: "this\n" + `"emptyBoolArray:[]; emptyFloat32Array:[]; emptyFloat64Array:[]; emptyInt16Array:[]; emptyInt32Array:[]; emptyInt64Array:[]; emptyInt8Array:[]; emptyNullableBoolArray:[]; emptyNullableInt8Array:[]; fullBoolArray:[true, false]; fullFloat32Array:[]; fullFloat64Array:[]; fullInt16Array:[-32768, 0, 32767]; fullInt32Array:[-2147483648, 0, 2147483647]; fullInt64Array:[-9223372036854775808, 0, -9223372036854775808]; fullInt8Array:[-128, 0, 127]; fullNullableBoolArray:[true]; fullNullableInt8Array:[-, 8]"` + "\n",
				},

				{
					format: "table",
					target: "testdata/primitive_arrays_table.txt",
				},
			}
			for _, tc := range testCases {
				tc := tc
				t.Run(tc.format, func(t *testing.T) {
					tcx.T = t
					tcx.WithReset(func() {
						check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "get", "value", "--quite", "-f", tc.format))
						if tc.format == "table" {
							tcx.AssertStdoutDollarWithPath(t, tc.target)
						} else {
							tcx.AssertStdoutEquals(t, tc.target)
						}
					})
				})
			}
		})

	})
}

func primitivesTest(t *testing.T) {
	tcx := it.TestContext{
		T: t,
		ConfigCallback: func(tcx it.TestContext) {
			tcx.ClientConfig.Serialization.Compact.SetSerializers(primitivesSerializer{})
		},
	}
	tcx.Tester(func(tcx it.TestContext) {
		i8 := int8(8)
		value := primitives{
			valueInt8:       i8,
			nullInt8NotNull: &i8,
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
					target: "nullInt8NotNull:8; nullInt8Null:-; valueInt8:8\n",
				},
				{
					format: "json",
					target: `{"this":{"nullInt8NotNull":8,"nullInt8Null":null,"valueInt8":8}}` + "\n",
				},

				{
					format: "csv",
					target: "this\n" + `nullInt8NotNull:8; nullInt8Null:-; valueInt8:8` + "\n",
				},
				{
					format: "table",
					target: "testdata/primitives_table.txt",
				},
			}
			for _, tc := range testCases {
				tc := tc
				t.Run(tc.format, func(t *testing.T) {
					tcx.T = t
					tcx.WithReset(func() {
						check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "get", "value", "--quite", "-f", tc.format))
						if tc.format == "table" {
							tcx.AssertStdoutDollarWithPath(t, tc.target)
						} else {
							tcx.AssertStdoutEquals(t, tc.target)
						}
					})
				})
			}
		})

	})
}

type primitives struct {
	valueInt8       int8
	nullInt8Null    *int8
	nullInt8NotNull *int8
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
	}
}

func (s primitivesSerializer) Write(w serialization.CompactWriter, v interface{}) {
	vv := v.(primitives)
	w.WriteInt8("valueInt8", vv.valueInt8)
	w.WriteNullableInt8("nullInt8Null", vv.nullInt8Null)
	w.WriteNullableInt8("nullInt8NotNull", vv.nullInt8NotNull)
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
