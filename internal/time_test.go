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
