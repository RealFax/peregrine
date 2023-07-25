package proto

import "io"

// Codec
//
// declares the impl required for serialization/deserialization
type Codec[T any, K comparable] interface {
	Marshal(w io.Writer, val Proto[T, K]) error
	Unmarshal(r io.Reader, ptr Proto[T, K]) error
}
