package texts

import (
	"reflect"
	"unsafe"
)

// ReadonlyString returns the string for the byte slice b.
func ReadonlyString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ReadOnlyBytes returns the backing byte slice for a string.
func ReadOnlyBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
