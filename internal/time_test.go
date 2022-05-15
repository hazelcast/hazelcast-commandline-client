package internal

import (
	"testing"
	"time"
)

func TestUserDuration_Validate(t *testing.T) {
	for _, tc := range []struct {
		msg      string
		isNotErr func(err error) bool
		in       time.Duration
	}{
		{msg: "zero", in: 0, isNotErr: func(err error) bool {
			return err != nil
		}},
		{msg: "equal to a second", in: time.Second, isNotErr: func(err error) bool {
			return err == nil
		}},
		{msg: "greater than a second", in: 2 * time.Second, isNotErr: func(err error) bool {
			return err == nil
		}},
		{msg: "less than a second as millisecond", in: 500 * time.Millisecond, isNotErr: func(err error) bool {
			return err != nil
		}},
		{msg: "greater than a second as minute", in: time.Minute, isNotErr: func(err error) bool {
			return err == nil
		}},
	} {
		t.Run(tc.msg, func(t *testing.T) {
			var err error
			d := &UserDuration{Duration: tc.in, DurType: TTL}
			err = d.Validate()
			if !tc.isNotErr(err) {
				t.Fatalf("error state is not satisfied")
			}
		})
	}
}
