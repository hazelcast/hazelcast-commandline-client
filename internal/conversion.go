package internal

import (
	"fmt"
	"strconv"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

// supported types
const (
	String  = "string"
	Boolean = "boolean"
	JSON    = "json"
	Int8    = "int8"
	Int16   = "int16"
	Int32   = "int32"
	Int64   = "int64"
	Float32 = "float32"
	Float64 = "float64"
)

var SupportedTypes = []string{
	String,
	Boolean,
	JSON,
	Int8,
	Int16,
	Int32,
	Int64,
	Float32,
	Float64,
}

func ConvertString(value, valueType string) (interface{}, error) {
	var (
		convertedValue interface{}
		err            error
	)
	var (
		intermediateInt   int64
		intermediateFloat float64
	)
	switch valueType {
	case String:
		convertedValue = value
	case Boolean:
		convertedValue, err = strconv.ParseBool(value)
	case JSON:
		convertedValue = serialization.JSON(value)
	case Int8:
		intermediateInt, err = strconv.ParseInt(value, 10, 8)
		convertedValue = int8(intermediateInt)
	case Int16:
		intermediateInt, err = strconv.ParseInt(value, 10, 16)
		convertedValue = int16(intermediateInt)
	case Int32:
		intermediateInt, err = strconv.ParseInt(value, 10, 32)
		convertedValue = int32(intermediateInt)
	case Int64:
		convertedValue, err = strconv.ParseInt(value, 10, 64)
	case Float32:
		intermediateFloat, err = strconv.ParseFloat(value, 32)
		convertedValue = float32(intermediateFloat)
	case Float64:
		convertedValue, err = strconv.ParseFloat(value, 64)
	default:
		err = fmt.Errorf("unknown type, provide one of %v", SupportedTypes)
	}
	if numErr, ok := err.(*strconv.NumError); ok {
		if numErr.Err == strconv.ErrSyntax {
			err = fmt.Errorf(`can not convert "%s" to %s, unknown syntax`, value, valueType)
		} else if numErr.Err == strconv.ErrRange {
			err = fmt.Errorf(`%s can not be represented with specified bit number (max val:%v)`, value, convertedValue)
		}
	}
	return convertedValue, err
}
