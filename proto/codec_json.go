package proto

import (
	"encoding/json"
	"io"
)

type CodecJSON[T any, K comparable] struct{}

func (k CodecJSON[T, K]) Marshal(_ string, w io.Writer, val Proto[T, K]) error {
	return json.NewEncoder(w).Encode(val.Value())
}

func (k CodecJSON[T, K]) Unmarshal(_ string, r io.Reader, ptr Proto[T, K]) error {
	return json.NewDecoder(r).Decode(ptr.Self())
}
