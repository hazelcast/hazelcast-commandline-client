package internal

import (
	"errors"
	"fmt"
	"time"
)

type DurationType string

const (
	TTL                  = "TTL"
	MaxIdle DurationType = "MaxIdle"
)

// UserDuration represents given custom duration value from the user
type UserDuration struct {
	time.Duration
	DurType DurationType
}

// Validate validates user duration type
func (d *UserDuration) Validate() error {
	if d.Seconds() < 0 {
		return errors.New(fmt.Sprintf("duration %s must be positive", d.DurType))
	}
	if d.DurType == MaxIdle {
		return nil
	}
	if d.DurType == TTL {
		if d.Seconds() >= time.Second.Seconds() {
			return nil
		}
		return errors.New("ttl duration cannot be less than a second")
	}
	return errors.New("undefined duration type")
}
