package format

import (
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"
)

// Fmt defines output format for different SQL types
func Fmt(v interface{}) string {
	switch t := v.(type) {
	case types.LocalDate:
		return time.Time(t).Format("2006-01-02")
	case types.LocalDateTime:
		return time.Time(t).Format("2006-01-02T15:04:05.999999")
	case types.LocalTime:
		return time.Time(t).Format("15:04:05.999999")
	case types.OffsetDateTime:
		return time.Time(t).Format(time.RFC3339)
	default:
		return fmt.Sprint(t)
	}
}
