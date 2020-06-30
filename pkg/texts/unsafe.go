package texts

import (
	"unsafe"
)

func ReadonlyStringSlice(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
