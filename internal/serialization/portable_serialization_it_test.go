package serialization_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/it"
)

func TestPortableSerialization(t *testing.T) {
	testCases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{name: "PortableArrays", f: portableArraysTest},
		{name: "PortableOthers", f: portableOthersTest},
		{name: "PortablePrimitives", f: portablePrimitivesTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func portablePrimitivesTest(t *testing.T) {
	testCases := []struct {
		format string
		target string
	}{
		{
			format: "delimited",
			target: "{valueBool:true; valueByte:8}\n",
		},
		{
			format: "json",
			target: `{"this":{"valueBool":true,"valueByte":8}}` + "\n",
		},
		{
			format: "csv",
			target: "this\n" + "{valueBool:true; valueByte:8}\n",
		},
		{
			format: "table",
			target: "testdata/portable_primitives_table.txt",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			portableTestValue(t, tc.format, tc.target, &primitives2{
				valueByte: 8,
				valueBool: true,
			})
		})
	}
}

func portableOthersTest(t *testing.T) {
	testCases := []struct {
		format string
		target string
	}{
		{
			format: "delimited",
			target: "{decimalNotNull:1.234E-53; decimalNull:-; offsetDateTimeNotNull:2023-04-05T12:33:45Z; offsetDateTimeNull:-; portable:{value:foo}; valueString:foobar}\n",
		},
		{
			format: "json",
			target: `{"this":{"decimalNotNull":"1.234E-53","decimalNull":null,"offsetDateTimeNotNull":"2023-04-05T12:33:45Z","offsetDateTimeNull":null,"portable":{"value":"foo"},"valueString":"foobar"}}` + "\n",
		},
		{
			format: "csv",
			target: "this\n" + `{decimalNotNull:1.234E-53; decimalNull:-; offsetDateTimeNotNull:2023-04-05T12:33:45Z; offsetDateTimeNull:-; portable:{value:foo}; valueString:foobar}` + "\n",
		},
		{
			format: "table",
			target: "testdata/portable_others_table.txt",
		},
	}
	dtz := time.Date(2023, 4, 5, 12, 33, 45, 46, time.UTC)
	dc := types.NewDecimal(big.NewInt(1234), 56)
	for _, tc := range testCases {
		portableTestValue(t, tc.format, tc.target, &portableOthers{
			valueString:           "foobar",
			offsetDateTimeNotNull: (*types.OffsetDateTime)(&dtz),
			decimalNotNull:        &dc,
			portable:              &simplePortable{value: "foo"},
		})
	}
}

func portableArraysTest(t *testing.T) {
	testCases := []struct {
		format string
		target string
	}{
		{
			format: "delimited",
			target: "{portableArray:[{value:value1}, {value:value2}]; stringArray:[str1, str2]}\n",
		},
		{
			format: "json",
			target: `{"this":{"portableArray":[{"value":"value1"},{"value":"value2"}],"stringArray":["str1","str2"]}}` + "\n",
		},
		{
			format: "csv",
			target: "this\n" + `"{portableArray:[{value:value1}, {value:value2}]; stringArray:[str1, str2]}"` + "\n",
		},
		{
			format: "table",
			target: "testdata/portable_arrays_table.txt",
		},
	}
	for _, tc := range testCases {
		portableTestValue(t, tc.format, tc.target, &portableArrays{
			stringArray: []string{"str1", "str2"},
			portableArray: []serialization.Portable{
				&simplePortable{value: "value1"},
				&simplePortable{value: "value2"},
			},
		})
	}
}

func portableTestValue(t *testing.T, format, target string, value any) {
	ctx := context.Background()
	t.Run(format, func(t *testing.T) {
		portableTester(t, func(tcx it.TestContext) {
			it.WithRandomMap(tcx, func(m *hazelcast.Map) {
				check.Must(m.Set(ctx, "value", value))
				tcx.WithReset(func() {
					ctx := context.Background()
					check.Must(tcx.CLC().Execute(ctx, "map", "-n", m.Name(), "get", "value", "-q", "-f", format))
					if format == "table" {
						tcx.AssertStdoutDollarWithPath(target)
					} else {
						tcx.AssertStdoutEquals(target)
					}
				})
			})
		})
	})
}

const (
	factoryID             = 1000
	othersClassID         = 10
	primitivesClassID     = 11
	simplePortableClassID = 12
	portableArraysClassID = 13
)

type simplePortable struct {
	value string
}

func (p *simplePortable) FactoryID() int32 {
	return factoryID
}

func (p *simplePortable) ClassID() int32 {
	return simplePortableClassID
}

func (p *simplePortable) WritePortable(w serialization.PortableWriter) {
	w.WriteString("value", p.value)
}

func (p *simplePortable) ReadPortable(r serialization.PortableReader) {
	p.value = r.ReadString("value")
}

type portableOthers struct {
	valueString           string
	offsetDateTimeNull    *types.OffsetDateTime
	offsetDateTimeNotNull *types.OffsetDateTime
	decimalNull           *types.Decimal
	decimalNotNull        *types.Decimal
	portable              serialization.Portable
}

func (o *portableOthers) FactoryID() int32 {
	return factoryID
}

func (o *portableOthers) ClassID() int32 {
	return othersClassID
}

func (o *portableOthers) WritePortable(w serialization.PortableWriter) {
	w.WriteString("valueString", o.valueString)
	w.WriteTimestampWithTimezone("offsetDateTimeNull", o.offsetDateTimeNull)
	w.WriteTimestampWithTimezone("offsetDateTimeNotNull", o.offsetDateTimeNotNull)
	w.WriteDecimal("decimalNull", o.decimalNull)
	w.WriteDecimal("decimalNotNull", o.decimalNotNull)
	w.WritePortable("portable", o.portable)
}

func (o *portableOthers) ReadPortable(r serialization.PortableReader) {
	o.valueString = r.ReadString("valueString")
	o.offsetDateTimeNull = r.ReadTimestampWithTimezone("offsetDateTimeNull")
	o.offsetDateTimeNotNull = r.ReadTimestampWithTimezone("offsetDateTimeNotNull")
	o.decimalNull = r.ReadDecimal("decimalNull")
	o.decimalNotNull = r.ReadDecimal("decimalNotNull")
	o.portable = r.ReadPortable("portable")
}

type primitives2 struct {
	valueByte byte
	valueBool bool
}

func (p *primitives2) FactoryID() int32 {
	return factoryID
}

func (p *primitives2) ClassID() int32 {
	return primitivesClassID
}

func (p *primitives2) WritePortable(w serialization.PortableWriter) {
	w.WriteByte("valueByte", p.valueByte)
	w.WriteBool("valueBool", p.valueBool)
}

func (p *primitives2) ReadPortable(r serialization.PortableReader) {
	p.valueByte = r.ReadByte("valueByte")
	p.valueBool = r.ReadBool("valueBool")
}

type portableArrays struct {
	stringArray   []string
	portableArray []serialization.Portable
}

func (p *portableArrays) FactoryID() int32 {
	return factoryID
}

func (p *portableArrays) ClassID() int32 {
	return portableArraysClassID
}

func (p *portableArrays) WritePortable(w serialization.PortableWriter) {
	w.WriteStringArray("stringArray", p.stringArray)
	w.WritePortableArray("portableArray", p.portableArray)
}

func (p *portableArrays) ReadPortable(r serialization.PortableReader) {
	p.stringArray = r.ReadStringArray("stringArray")
	p.portableArray = r.ReadPortableArray("portableArray")
}

type portableFactory struct{}

func (p portableFactory) Create(cid int32) serialization.Portable {
	switch cid {
	case othersClassID:
		return &portableOthers{}
	case primitivesClassID:
		return &primitives2{}
	case simplePortableClassID:
		return &simplePortable{}
	case portableArraysClassID:
		return &portableArrays{}
	}
	panic(fmt.Sprintf("unknown cid: %d", cid))
}

func (p portableFactory) FactoryID() int32 {
	return factoryID
}

func portableTester(t *testing.T, f func(tcx it.TestContext)) {
	tcx := it.TestContext{
		T: t,
		ConfigCallback: func(tcx it.TestContext) {
			tcx.ClientConfig.Serialization.SetPortableFactories(&portableFactory{})
		},
	}
	tcx.Tester(f)
}
