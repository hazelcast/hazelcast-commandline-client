package serialization

import (
	"fmt"
	"time"

	"github.com/hazelcast/hazelcast-go-client/types"

	"github.com/hazelcast/hazelcast-commandline-client/errors"
)

const (
	ValueUnknown    = "UNKNOWN"
	ValueNotDecoded = "*"
	ValueSkip       = ">"
	ValueNil        = "-"
)

func MarshalLocalDate(v any) (*string, error) {
	switch vv := v.(type) {
	case types.LocalDate:
		s := time.Time(vv).Format("2006-01-02")
		return &s, nil
	case *types.LocalDate:
		if vv == nil {
			return nil, nil
		}
		s := (*time.Time)(vv).Format("2006-01-02")
		return &s, nil
	default:
		return nil, errors.ErrNotDecoded
	}
}

func MarshalLocalTime(v any) (*string, error) {
	switch vv := v.(type) {
	case types.LocalTime:
		s := time.Time(vv).Format("15:04:05")
		return &s, nil
	case *types.LocalTime:
		if vv == nil {
			return nil, nil
		}
		s := (*time.Time)(vv).Format("15:04:05")
		return &s, nil
	default:
		return nil, errors.ErrNotDecoded
	}
}

func MarshalLocalDateTime(v any) (*string, error) {
	switch vv := v.(type) {
	case types.LocalDateTime:
		s := time.Time(vv).Format("2006-01-02 15:04:05")
		return &s, nil
	case *types.LocalDateTime:
		if vv == nil {
			return nil, nil
		}
		s := (*time.Time)(vv).Format("2006-01-02 15:04:05")
		return &s, nil
	default:
		return nil, errors.ErrNotDecoded
	}
}

func MarshalOffsetDateTime(v any) (*string, error) {
	switch vv := v.(type) {
	case types.OffsetDateTime:
		s := time.Time(vv).Format(time.RFC3339)
		return &s, nil
	case *types.OffsetDateTime:
		if vv == nil {
			return nil, nil
		}
		s := (*time.Time)(vv).Format(time.RFC3339)
		return &s, nil
	default:
		return nil, errors.ErrNotDecoded
	}
}

func MarshalDecimal(v any) (string, error) {
	switch vv := v.(type) {
	case types.Decimal:
		return fmt.Sprintf("%s/^%d", vv.UnscaledValue().String(), vv.Scale()), nil
	case *types.Decimal:
		if vv == (*types.Decimal)(nil) {
			return ValueNil, nil
		}
		return fmt.Sprintf("%s/^%d", vv.UnscaledValue().String(), vv.Scale()), nil
	default:
		return "", errors.ErrNotDecoded
	}
}
