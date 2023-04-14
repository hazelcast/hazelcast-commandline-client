package check

import "reflect"

func IsNil(v any) bool {
	if v == nil {
		return true
	}
	vv := reflect.ValueOf(v)
	switch vv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map,
		reflect.Pointer, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice:
		return vv.IsNil()
	}
	return false
}
