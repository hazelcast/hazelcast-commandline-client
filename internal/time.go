/*
 * Copyright (c) 2008-2021, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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
		if d.Seconds() >= 1.0 {
			return nil
		}
		return errors.New("ttl duration cannot be less than a second")
	}
	return errors.New("undefined duration type")
}
