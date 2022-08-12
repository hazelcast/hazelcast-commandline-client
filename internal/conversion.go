package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hazelcast/hazelcast-go-client/serialization"
)

// supported types
const (
	TypeNameString  = "string"
	TypeNameBoolean = "bool"
	TypeNameJSON    = "json"
	TypeNameInt8    = "int8"
	TypeNameInt16   = "int16"
	TypeNameInt32   = "int32"
	TypeNameInt64   = "int64"
	TypeNameFloat32 = "float32"
	TypeNameFloat64 = "float64"
)

var SupportedTypeNames = []string{
	TypeNameString,
	TypeNameBoolean,
	TypeNameJSON,
	TypeNameInt8,
	TypeNameInt16,
	TypeNameInt32,
	TypeNameInt64,
	TypeNameFloat32,
	TypeNameFloat64,
}

func ConvertString(value, valueType string) (interface{}, error) {
	var (
		cv  interface{}
		err error
		i   int64
		f   float64
	)
	valueType = strings.ToLower(valueType)
	switch valueType {
	// "" is for default/empty
	case TypeNameString, "":
		cv = value
	case TypeNameBoolean:
		cv, err = strconv.ParseBool(value)
		if value == "true" {
			cv = true
		} else if value == "false" {
			cv = false
		} else {
			err = errors.New("bool values can be only true or false")
		}
	case TypeNameJSON:
		if !json.Valid([]byte(value)) {
			err = errors.New("malformed JSON string")
			break
		}
		cv = serialization.JSON(value)
	case TypeNameInt8:
		i, err = strconv.ParseInt(value, 10, 8)
		cv = int8(i)
	case TypeNameInt16:
		i, err = strconv.ParseInt(value, 10, 16)
		cv = int16(i)
	case TypeNameInt32:
		i, err = strconv.ParseInt(value, 10, 32)
		cv = int32(i)
	case TypeNameInt64:
		cv, err = strconv.ParseInt(value, 10, 64)
	case TypeNameFloat32:
		f, err = strconv.ParseFloat(value, 32)
		cv = float32(f)
	case TypeNameFloat64:
		cv, err = strconv.ParseFloat(value, 64)
	default:
		err = fmt.Errorf("unknown type, provide one of %s", strings.Join(SupportedTypeNames, ","))
	}
	if errors.Is(err, strconv.ErrSyntax) {
		err = fmt.Errorf(`can not convert "%s" to %s, unknown syntax`, value, valueType)
	} else if errors.Is(err, strconv.ErrRange) {
		err = fmt.Errorf(`%s can not be represented with specified bit number (max val:%v)`, value, cv)
	}
	return cv, err
}
