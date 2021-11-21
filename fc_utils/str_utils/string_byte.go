package str_utils

import "unsafe"

func Str2Bytes(res string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&res))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
