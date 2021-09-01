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
)

type Error struct {
	Short string
	Long  string
	Err   error
}

func (e *Error) AddErr(err error) *Error {
	e.Err = err
	return e
}

func (e *Error) NonVerboseErrorOut() string {
	return fmt.Sprintf(`
		mode: default,
		definition: %s
	`, e.Short)
}

func (e *Error) VerboseErrorOut() string {
	return fmt.Sprintf(`
		mode: verbose,
		definition: %s,
		error: %s
	`, e.Long, e.Err)
}

var ErrMapNameMissing = Error{
	Short: "MapNameMissing",
	Long:  "map name is required",
	Err:   nil,
}

var ErrMapKeyMissing = Error{
	Short: "MapKeyMissingError",
	Long:  "map key is required",
	Err:   nil,
}

var ErrMapValueMissing = Error{
	Short: "MapValueMissing",
	Long:  "map value is required.",
	Err:   nil,
}

var ErrMapValueAndFileMutuallyExclusive = Error{
	Short: "MapValueAndFileMutuallyExclusive",
	Long:  "only one of --value and --value-file must be specified",
	Err:   nil,
}

var ErrMapUnableToBeRetrieved = Error{
	Short: "MapUnableToBeRetrievedError",
	Long:  "map is unable to be retrieved after a successful connection",
	Err:   nil,
}

var ErrMapValueInvalid = Error{
	Short: "MapValueInvalidError",
	Long:  "map value is invalid to be retrieved",
	Err:   nil,
}

var ErrMapValueUnsuccessfulPut = Error{
	Short: "MapValueUnsuccessfulPutError",
	Long:  "map value is unable to be put into the map",
	Err:   nil,
}

var ErrMapValueFileLoadInvalid = Error{
	Short: "MapValueFileLoadInvalidError",
	Long:  "map value can not be loaded from file",
	Err:   nil,
}

var ErrMapValueTypeInvalid = Error{
	Short: "MapValueTypeInvalidError",
	Long:  "given map value type is invalid",
	Err:   nil,
}

var ErrEmptyFilePath = Error{
	Short: "FilePathEmptyError",
	Long:  "provided file path can not be empty",
	Err:   nil,
}

var ErrConfigRead = Error{
	Short: "ConfigReadError",
	Long:  "config can not be read",
	Err:   nil,
}

var ErrConfigWrite = Error{
	Short: "ConfigWriteError",
	Long:  "config can not be written",
	Err:   nil,
}

var ErrConfigUnmarshalInvalid = Error{
	Short: "ConfigInvalidError",
	Long:  "invalid config to be unmarshalled",
	Err:   nil,
}

var ErrConfigInvalid = Error{
	Short: "ConfigInvalidError",
	Long:  "config is invalid",
	Err:   nil,
}

var ErrAddressInvalid = Error{
	Short: "AddressInvalidError",
	Long:  "invalid address is provided",
	Err:   nil,
}

var ErrClientNotCreated = Error{
	Short: "ClientNotCreatedError",
	Long:  "client can not be created",
	Err:   nil,
}

var ErrConnectionDeadlineExceeded = Error{
	Short: "ConnectionDeadlineExceededError",
	Long:  "connection deadline exceeded due to unsuccessful connection",
	Err:   nil,
}

var ErrClusterGetStateOperation = Error{
	Short: "ClusterGetStateOperationError",
	Long:  "failed to perform cluster get state operation",
	Err:   nil,
}

var ErrClusterChangeStateOperation = Error{
	Short: "ClusterChangeStateOperationError",
	Long:  "failed to perform cluster change state operation",
	Err:   nil,
}

var ErrClusterShutdownOperation = Error{
	Short: "ClusterShutdownOperationError",
	Long:  "failed to perform cluster shutdown operation",
	Err:   nil,
}

var ErrClusterVersionOperation = Error{
	Short: "ClusterVersionOperationError",
	Long:  "failed to perform cluster version operation",
	Err:   nil,
}

var ErrInvalidNewState = Error{
	Short: "InvalidNewStateError",
	Long:  "given new state is invalid. It can be either of: { active | frozen | no_migration | passive}",
	Err:   nil,
}

var ErrInvalidClusterOperation = Error{
	Short: "InvalidClusterOperationError",
	Long:  "target cluster operation is invalid",
	Err:   nil,
}

func (e *Error) NonVerboseError() error {
	return errors.New(e.Short)
}

func (e *Error) VerboseError() error {
	return errors.New(e.Long)
}

func ErrorRecover() {
	obj := recover()
	if err, ok := obj.(error); ok {
		fmt.Println(err)
	}
}
