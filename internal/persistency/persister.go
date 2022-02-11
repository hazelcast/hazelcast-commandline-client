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
package persister

func NewNamePersister() *namePersistence {
	var fp namePersistence
	fp.names = make(map[string]string)
	return &fp
}

type namePersistence struct {
	names map[string]string
}

// Set sets the value to persist
func (f *namePersistence) Set(name string, value string) {
	f.names[name] = value
}

// Get returns the set value for the name. Second argument returns false if there is no value set.
func (f *namePersistence) Get(name string) (string, bool) {
	val, ok := f.names[name]
	return val, ok
}

// Reset clears the set value for the name
func (f *namePersistence) Reset(name string) {
	delete(f.names, name)
}

// PersistenceInfo returns stored values
func (f *namePersistence) PersistenceInfo() map[string]string {
	return f.names
}
