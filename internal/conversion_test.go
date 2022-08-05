package internal

import (
	"testing"

	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/stretchr/testify/require"
)

func TestConvertString(t *testing.T) {
	type args struct {
		value     string
		valueType string
	}
	tests := []struct {
		name  string
		args  args
		want  interface{}
		isErr bool
	}{
		{
			name: "valid string",
			args: args{
				value:     "abc",
				valueType: TypeNameString,
			},
			want: "abc",
		},
		{
			name: "valid Boolean",
			args: args{
				value:     "1",
				valueType: TypeNameBoolean,
			},
			want:  true,
			isErr: false,
		},
		{
			name: "valid Boolean",
			args: args{
				value:     "true",
				valueType: TypeNameBoolean,
			},
			want:  true,
			isErr: false,
		},
		{
			name: "valid Boolean",
			args: args{
				value:     "0",
				valueType: TypeNameBoolean,
			},
			want:  false,
			isErr: false,
		},
		{
			name: "valid Boolean",
			args: args{
				value:     "FALSE",
				valueType: TypeNameBoolean,
			},
			want:  false,
			isErr: false,
		},
		{
			name: "valid JSON",
			args: args{
				value:     `{"test":"data"}`,
				valueType: TypeNameJSON,
			},
			want: serialization.JSON(`{"test":"data"}`),
		},
		{
			name: "valid JSON",
			args: args{
				value:     `"jsonString"`,
				valueType: TypeNameJSON,
			},
			want: serialization.JSON("jsonString"),
		},
		{
			name: "valid JSON",
			args: args{
				value:     `[1,2,3,]`,
				valueType: TypeNameJSON,
			},
			isErr: true,
		},
		{
			name: "valid Int8",
			args: args{
				value:     "127",
				valueType: TypeNameInt8,
			},
			want: int8(127),
		},
		{
			name: "valid Int8",
			args: args{
				value:     "-128",
				valueType: TypeNameInt8,
			},
			want: int8(-128),
		},
		{
			name: "invalid Int8, overflow",
			args: args{
				value:     "128",
				valueType: TypeNameInt8,
			},
			isErr: true,
		},
		{
			name: "valid Int16",
			args: args{
				value:     "12345",
				valueType: TypeNameInt16,
			},
			want: int16(12345),
		},
		{
			name: "valid Int32",
			args: args{
				value:     "12345",
				valueType: TypeNameInt32,
			},
			want: int32(12345),
		},
		{
			name: "valid Int64",
			args: args{
				value:     "12345",
				valueType: TypeNameInt64,
			},
			want: int64(12345),
		},
		{
			name: "valid Float32",
			args: args{
				value:     "0.34",
				valueType: TypeNameFloat32,
			},
			want: float32(0.34),
		},
		{
			name: "valid Float64",
			args: args{
				value:     "0.34",
				valueType: TypeNameFloat64,
			},
			want: float64(0.34),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertString(tt.args.value, tt.args.valueType)
			if (err != nil) != tt.isErr {
				t.FailNow()
				return
			}
			if err != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
