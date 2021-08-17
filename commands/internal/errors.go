package internal

import (
	"errors"
	"fmt"
)

var ErrMapKeyMissing = errors.New("map key is required")
var ErrMapValueMissing = errors.New("map value is required")
var ErrMapValueAndFileMutuallyExclusive = errors.New("only one of --value and --value-file must be specified")

func HzDefer() {
	obj := recover()
	if err, ok := obj.(error); ok {
		fmt.Println(err)
	}
}
