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
package verbose

import (
	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

func Resolve(isVerbose bool, color color.Color, output interface{}, err *internal.Error) string {
	if isVerbose {
		if err != nil {
			return color.Sprintf(err.VerboseErrorOut(), color)
		}
		return color.Sprintf("%v", output)
	}
	if err != nil {
		return color.Sprintf(err.NonVerboseErrorOut(), color)
	}
	return color.Sprintf("%v", output)
}
