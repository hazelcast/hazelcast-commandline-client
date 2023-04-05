package serialization

import (
	"testing"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"
	"github.com/stretchr/testify/require"
)

func TestTimeStringer(t *testing.T) {
	var nullTm *time.Time
	tm := time.Date(2023, 1, 2, 3, 4, 5, 6, time.UTC)
	testCases := []struct {
		name   string
		in     any
		target string
	}{
		{
			name:   "null time.Time",
			in:     nullTm,
			target: ValueNil,
		},
		{
			name:   "time.Time",
			in:     tm,
			target: "2023-01-02T03:04:05Z",
		},
		{
			name:   "null *types.LocalTime",
			in:     (*types.LocalTime)(nullTm),
			target: ValueNil,
		},
		{
			name:   "*types.LocalTime",
			in:     (*types.LocalTime)(&tm),
			target: "03:04:05",
		},
		{
			name:   "types.LocalTime",
			in:     (types.LocalTime)(tm),
			target: "03:04:05",
		},
		{
			name:   "null *types.LocalDate",
			in:     (*types.LocalDate)(nullTm),
			target: ValueNil,
		},
		{
			name:   "*types.LocalDate",
			in:     (*types.LocalDate)(&tm),
			target: "2023-01-02",
		},
		{
			name:   "types.LocalDate",
			in:     (types.LocalDate)(tm),
			target: "2023-01-02",
		},
		{
			name:   "null *types.LocalDateTime",
			in:     (*types.LocalDateTime)(nullTm),
			target: ValueNil,
		},
		{
			name:   "*types.LocalDateTime",
			in:     (*types.LocalDateTime)(&tm),
			target: "2023-01-02 03:04:05",
		},
		{
			name:   "types.LocalDateTime",
			in:     (types.LocalDateTime)(tm),
			target: "2023-01-02 03:04:05",
		},
		{
			name:   "null *types.OffsetDateTime",
			in:     (*types.OffsetDateTime)(nullTm),
			target: ValueNil,
		},
		{
			name:   "*types.OffsetDateTime",
			in:     (*types.OffsetDateTime)(&tm),
			target: "2023-01-02T03:04:05Z",
		},
		{
			name:   "types.OffsetDateTime",
			in:     (types.OffsetDateTime)(tm),
			target: "2023-01-02T03:04:05Z",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			v := timeStringer(tc.in)
			require.Equal(t, tc.target, v)
		})
	}
}
