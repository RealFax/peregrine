package proto

import "sync"

type InstancePool[T any, K comparable] struct {
	pool sync.Pool
}

func (k *InstancePool[T, K]) Alloc() Proto[T, K] {
	return k.pool.Get().(Proto[T, K])
}

func (k *InstancePool[T, K]) Free(proto Proto[T, K]) {
	// if Proto impl Resetter interface
	if free, ok := proto.(Resetter); ok {
		free.Reset()
	}
	k.pool.Put(proto)
}

func NewInstancePool[T any, K comparable](newProto func() Proto[T, K]) *InstancePool[T, K] {
	return &InstancePool[T, K]{
		pool: sync.Pool{
			New: func() any { return newProto() },
		},
	}
}
