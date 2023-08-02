package proto

import "io"

// Codec
//
// declares the impl required for serialization/deserialization
type Codec[T any, K comparable] interface {
	Marshal(connID string, w io.Writer, val Proto[T, K]) error
	Unmarshal(connID string, r io.Reader, ptr Proto[T, K]) error
}
