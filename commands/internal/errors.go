package internal

import (
	"errors"
	"fmt"
)

var ErrMapKeyMissing = errors.New("map key is required")
var ErrMapValueMissing = errors.New("map value is required")
var ErrMapValueAndFileMutuallyExclusive = errors.New("only one of --value and --value-file must be specified")

func RaiseErrAddressInvalid(addr string) error {
	return fmt.Errorf("given address is invalid: %s", addr)
}
