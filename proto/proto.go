package proto

type Proto[T any, K comparable] interface {
	Key() K
	Value() T
	Self() *T
}

func New[T any, K comparable](newProto func() Proto[T, K]) *Engine[T, K] {
	return &Engine[T, K]{
		codec:    &CodecJSON[T, K]{},
		handlers: make(map[K]HandlerFunc[T]),
		newProto: newProto,
	}
}
