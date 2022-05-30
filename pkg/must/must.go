package must

import "net/http"

// OK panics if err is not nil
func OK(err error) {
	if err != nil {
		panic(err)
	}
}

// Bool panics if err is not nil, v is returned otherwise
func Bool(v bool, err error) bool {
	OK(err)
	return v
}

// String panics if err is not nil, v is returned otherwise
func String(v string, err error) string {
	OK(err)
	return v
}

// Bytes panics if err is not nil, v is returned otherwise
func Bytes(v []byte, err error) []byte {
	OK(err)
	return v
}

// Int panics if err is not nil, v is returned otherwise
func Int(v int, err error) int {
	OK(err)
	return v
}

// Int8 panics if err is not nil, v is returned otherwise
func Int8(v int8, err error) int8 {
	OK(err)
	return v
}

// Int16 panics if err is not nil, v is returned otherwise
func Int16(v int16, err error) int16 {
	OK(err)
	return v
}

// Int32 panics if err is not nil, v is returned otherwise
func Int32(v int32, err error) int32 {
	OK(err)
	return v
}

// Int64 panics if err is not nil, v is returned otherwise
func Int64(v int64, err error) int64 {
	OK(err)
	return v
}

// UInt panics if err is not nil, v is returned otherwise
func UInt(v uint, err error) uint {
	OK(err)
	return v
}

// UInt8 panics if err is not nil, v is returned otherwise
func UInt8(v uint8, err error) uint8 {
	OK(err)
	return v
}

// UInt16 panics if err is not nil, v is returned otherwise
func UInt16(v uint16, err error) uint16 {
	OK(err)
	return v
}

// UInt32 panics if err is not nil, v is returned otherwise
func UInt32(v uint32, err error) uint32 {
	OK(err)
	return v
}

// UInt64 panics if err is not nil, v is returned otherwise
func UInt64(v uint64, err error) uint64 {
	OK(err)
	return v
}

// Any panics if err is not ni
func Any(_ interface{}, err error) {
	OK(err)
}

// HTTPRequest panics if err is not nil, v is returned otherwise
func HTTPRequest(v *http.Request, err error) *http.Request {
	OK(err)
	return v
}
