package hack

import "unsafe"

func Bytes2String(b []byte) string {
	return unsafe.String(&b[0], len(b))
}
