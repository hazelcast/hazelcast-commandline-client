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
	tcs := []struct {
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
			name: "invalid boolean",
			args: args{
				value:     "1",
				valueType: TypeNameBoolean,
			},
			isErr: true,
		},
		{
			name: "valid boolean",
			args: args{
				value:     "true",
				valueType: TypeNameBoolean,
			},
			want:  true,
			isErr: false,
		},
		{
			name: "invalid boolean",
			args: args{
				value:     "FALSE",
				valueType: TypeNameBoolean,
			},
			isErr: true,
		},
		{
			name: "valid json",
			args: args{
				value:     `{"test":"data"}`,
				valueType: TypeNameJSON,
			},
			want: serialization.JSON(`{"test":"data"}`),
		},
		{
			name: "valid json",
			args: args{
				value:     `"jsonString"`,
				valueType: TypeNameJSON,
			},
			want: serialization.JSON(`"jsonString"`),
		},
		{
			name: "valid json",
			args: args{
				value:     `[1,2,3,]`,
				valueType: TypeNameJSON,
			},
			isErr: true,
		},
		{
			name: "valid int8",
			args: args{
				value:     "127",
				valueType: TypeNameInt8,
			},
			want: int8(127),
		},
		{
			name: "valid int8",
			args: args{
				value:     "-128",
				valueType: TypeNameInt8,
			},
			want: int8(-128),
		},
		{
			name: "invalid int8, overflow",
			args: args{
				value:     "128",
				valueType: TypeNameInt8,
			},
			isErr: true,
		},
		{
			name: "valid int16",
			args: args{
				value:     "12345",
				valueType: TypeNameInt16,
			},
			want: int16(12345),
		},
		{
			name: "valid int32",
			args: args{
				value:     "12345",
				valueType: TypeNameInt32,
			},
			want: int32(12345),
		},
		{
			name: "valid int64",
			args: args{
				value:     "12345",
				valueType: TypeNameInt64,
			},
			want: int64(12345),
		},
		{
			name: "valid float32",
			args: args{
				value:     "0.34",
				valueType: TypeNameFloat32,
			},
			want: float32(0.34),
		},
		{
			name: "valid float64",
			args: args{
				value:     "0.34",
				valueType: TypeNameFloat64,
			},
			want: float64(0.34),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConvertString(tc.args.value, tc.args.valueType)
			if (err != nil) != tc.isErr {
				t.Fatalf("expected and error")
				return
			}
			if err != nil {
				return
			}
			require.Equal(t, tc.want, got)
		})
	}
}
