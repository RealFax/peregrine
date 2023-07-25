package hack

import "unsafe"

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String2Bytes(s string) []byte {
	a := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{a[0], a[1], a[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}
