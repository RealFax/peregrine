package proto

import (
	"encoding/xml"
	"io"
)

type CodecXML[T any, K comparable] struct{}

func (k CodecXML[T, K]) Marshal(w io.Writer, val Proto[T, K]) error {
	return xml.NewEncoder(w).Encode(val.Value())
}

func (k CodecXML[T, K]) Unmarshal(r io.Reader, ptr Proto[T, K]) error {
	return xml.NewDecoder(r).Decode(ptr.Self())
}
