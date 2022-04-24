package util

import (
	"fmt"
	"reflect"
	"strconv"
)

// IsNotZero just IsZero invert
func IsNotZero(data interface{}) bool {
	return !IsZero(data)
}

// IsZero check data is zero value
// be care of the slice, a slice like []int{} is not zero
// because in Go, slice pointed to the underlying array,
// so slice is not zero.
func IsZero(data interface{}) bool {
	if data == nil {
		return true
	}
	if val, ok := data.(reflect.Value); ok {
		return val.IsZero()
	}
	return reflect.ValueOf(data).IsZero()
}

func StrArr2IntArr(s []string, ignoreZero bool) ([]int, error) {
	result := make([]int, 0, len(s))
	for i := range s {
		if len(s[i]) > 0 {
			v, err := strconv.ParseInt(s[i], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("s has wrong item in %d ,%w", i, err)
			}
			result = append(result, int(v))
		} else if !ignoreZero {
			result = append(result, 0)
		}
	}
	return result, nil
}
