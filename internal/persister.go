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

import "context"

func PersisterFromContext(ctx context.Context) NamePersister {
	return ctx.Value("persister").(NamePersister)
}

func ContextWithPersister(ctx context.Context, persister NamePersister) context.Context {
	return context.WithValue(ctx, "persister", persister)
}

func NewNamePersister() NamePersister {
	var nm NamePersister
	nm = make(map[string]string)
	return nm
}

type NamePersister map[string]string

// Set sets the value to persist
func (nm NamePersister) Set(name string, value string) {
	nm[name] = value
}

// Get returns the set value for the name. Second argument returns false if there is no value set.
func (nm NamePersister) Get(name string) (string, bool) {
	val, ok := nm[name]
	return val, ok
}

// Reset clears the set value for the name
func (nm NamePersister) Reset(name string) {
	delete(nm, name)
}

// PersistenceInfo returns stored values
func (nm NamePersister) PersistenceInfo() map[string]string {
	return nm
}
