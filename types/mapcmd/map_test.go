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

package mapcmd

import (
	"bytes"
	"testing"
	"time"
)

func TestObtainOrderingOfValues(t *testing.T) {
	for _, tc := range []struct {
		info string
		want []byte
		args []string
	}{
		{"value-short then file", []byte{'s', 'f'}, []string{"-v", "--value-file"}},
		{"file then value-short", []byte{'f', 's'}, []string{"--value-file", "-v"}},
		{"file-short then value", []byte{'f', 's'}, []string{"-f", "--value"}},
		{"value then file-short", []byte{'s', 'f'}, []string{"--value", "-f"}},
		{"value-short then file-short", []byte{'s', 'f'}, []string{"-v", "-f"}},
		{"file-short then value-short", []byte{'f', 's'}, []string{"-f", "-v"}},
		{"empty", nil, []string{}},
	} {
		t.Run(tc.info, func(t *testing.T) {
			gotvOrder, _ := ObtainOrderingOfValueFlags(tc.args)
			if !bytes.Equal(tc.want, gotvOrder) {
				t.Errorf("want %v got %v", tc.want, gotvOrder)
			}
		})
	}
}

func TestValidateTTL(t *testing.T) {
	for _, tc := range []struct {
		msg   string
		isErr bool
		in    time.Duration
	}{
		{msg: "zero", in: 0, isErr: true},
		{msg: "equal to a second", in: time.Second, isErr: false},
		{msg: "greater than a second", in: 2 * time.Second, isErr: false},
		{msg: "less than a second as millisecond", in: 500 * time.Millisecond, isErr: true},
		{msg: "greater than a second as minute", in: time.Minute, isErr: false},
	} {
		t.Run(tc.msg, func(t *testing.T) {
			err := validateTTL(tc.in)
			if (err != nil) != tc.isErr {
				t.Fatalf("error state is not satisfied")
			}
		})
	}
}
