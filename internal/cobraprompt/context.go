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
package cobraprompt

import (
	"context"
	"time"
)

// CobraMutableCtx is a workaround for https://github.com/spf13/cobra/issues/1109
type CobraMutableCtx struct {
	Internal context.Context
}

func (c *CobraMutableCtx) Deadline() (deadline time.Time, ok bool) {
	return c.Internal.Deadline()
}
func (c *CobraMutableCtx) Done() <-chan struct{} {
	return c.Internal.Done()
}
func (c *CobraMutableCtx) Err() error {
	return c.Internal.Err()
}
func (c *CobraMutableCtx) Value(key interface{}) interface{} {
	return c.Internal.Value(key)
}
