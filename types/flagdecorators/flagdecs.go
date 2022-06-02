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
package flagdecorators

import (
	"time"

	"github.com/spf13/cobra"
)

// common flags
const (
	JsonEntryFlag = "json-entry"
	TTLFlag       = "ttl"
	MaxIdleFlag   = "max-idle"
	DelimiterFlag = "delim"
)

func DecorateCommandWithJsonEntryFlag(cmd *cobra.Command, jsonEntry *string, required bool, usage string) {
	cmd.Flags().StringVar(jsonEntry, JsonEntryFlag, "", usage)
	if required {
		if err := cmd.MarkFlagRequired(JsonEntryFlag); err != nil {
			panic(err)
		}
	}
}

func DecorateCommandWithTTL(cmd *cobra.Command, ttl *time.Duration, required bool, usage string) {
	cmd.Flags().DurationVar(ttl, TTLFlag, 0, usage)
	if required {
		if err := cmd.MarkFlagRequired(TTLFlag); err != nil {
			panic(err)
		}
	}
}

func DecorateCommandWithMaxIdle(cmd *cobra.Command, maxIdle *time.Duration, required bool, usage string) {
	cmd.Flags().DurationVar(maxIdle, MaxIdleFlag, 0, usage)
	if required {
		if err := cmd.MarkFlagRequired(MaxIdleFlag); err != nil {
			panic(err)
		}
	}
}

func DecorateCommandWithDelimiter(cmd *cobra.Command, delimiter *string, required bool, usage string) {
	cmd.Flags().StringVar(delimiter, DelimiterFlag, "\t", usage)
	if required {
		if err := cmd.MarkFlagRequired(DelimiterFlag); err != nil {
			panic(err)
		}
	}
}
