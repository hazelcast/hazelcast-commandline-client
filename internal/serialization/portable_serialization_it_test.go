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
		{name: "PortableOthers", f: portableOthersTest},
		{name: "PortablePrimitives", f: portablePrimitivesTest},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.f)
	}
}

func portablePrimitivesTest(t *testing.T) {
	portableTester(t, func(tcx it.TestContext) {
		value := &primitives2{
			valueByte: 8,
			valueBool: true,
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
					target: "valueBool:true; valueByte:8\n",
				},
				{
					format: "json",
					target: `{"this":{"valueBool":true,"valueByte":8}}` + "\n",
				},
				{
					format: "csv",
					target: "this\n" + "valueBool:true; valueByte:8\n",
				},
				{
					format: "table",
					target: "testdata/portable_primitives_table.txt",
				},
			}
			for _, tc := range testCases {
				tc := tc
				t.Run(tc.format, func(t *testing.T) {
					tcx.T = t
					tcx.WithReset(func() {
						check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "get", "value", "-q", "-f", tc.format))
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

func portableOthersTest(t *testing.T) {
	portableTester(t, func(tcx it.TestContext) {
		dtz := time.Date(2023, 4, 5, 12, 33, 45, 46, time.UTC)
		dc := types.NewDecimal(big.NewInt(1234), 56)
		value := &others2{
			valueString:           "foobar",
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
					target: "decimalNotNull:1.234E-53; decimalNull:-; offsetDateTimeNotNull:2023-04-05T12:33:45Z; offsetDateTimeNull:-; valueString:foobar\n",
				},
				{
					format: "json",
					target: `{"this":{"decimalNotNull":"1.234E-53","decimalNull":null,"offsetDateTimeNotNull":"2023-04-05T12:33:45Z","offsetDateTimeNull":null,"valueString":"foobar"}}` + "\n",
				},
				{
					format: "csv",
					target: "this\n" + `decimalNotNull:1.234E-53; decimalNull:-; offsetDateTimeNotNull:2023-04-05T12:33:45Z; offsetDateTimeNull:-; valueString:foobar` + "\n",
				},
				{
					format: "table",
					target: "testdata/portable_others_table.txt",
				},
			}
			for _, tc := range testCases {
				tc := tc
				t.Run(tc.format, func(t *testing.T) {
					tcx.T = t
					tcx.WithReset(func() {
						check.Must(tcx.CLC().Execute("map", "-n", m.Name(), "get", "value", "-q", "-f", tc.format))
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

const (
	factoryID         = 1000
	othersClassID     = 10
	primitivesClassID = 11
)

// duplicated the others to be able to serialize in portable

type others2 struct {
	valueString           string
	offsetDateTimeNull    *types.OffsetDateTime
	offsetDateTimeNotNull *types.OffsetDateTime
	decimalNull           *types.Decimal
	decimalNotNull        *types.Decimal
}

func (o *others2) FactoryID() int32 {
	return factoryID
}

func (o *others2) ClassID() int32 {
	return othersClassID
}

func (o *others2) WritePortable(w serialization.PortableWriter) {
	w.WriteString("valueString", o.valueString)
	w.WriteTimestampWithTimezone("offsetDateTimeNull", o.offsetDateTimeNull)
	w.WriteTimestampWithTimezone("offsetDateTimeNotNull", o.offsetDateTimeNotNull)
	w.WriteDecimal("decimalNull", o.decimalNull)
	w.WriteDecimal("decimalNotNull", o.decimalNotNull)
}

func (o *others2) ReadPortable(r serialization.PortableReader) {
	o.valueString = r.ReadString("valueString")
	o.offsetDateTimeNull = r.ReadTimestampWithTimezone("offsetDateTimeNull")
	o.offsetDateTimeNotNull = r.ReadTimestampWithTimezone("offsetDateTimeNotNull")
	o.decimalNull = r.ReadDecimal("decimalNull")
	o.decimalNotNull = r.ReadDecimal("decimalNotNull")
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

type portableFactory struct{}

func (p portableFactory) Create(cid int32) serialization.Portable {
	switch cid {
	case othersClassID:
		return &others2{}
	case primitivesClassID:
		return &primitives2{}
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
