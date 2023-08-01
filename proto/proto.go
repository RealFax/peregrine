package proto

type Proto[T any, K comparable] interface {
	// Key realizes the value that returns Proto can be comparable
	//
	// case:
	// type Proto struct {
	//		Type 	uint32
	//		Message string
	// }
	//
	// func (p *Proto) Key() uint32 { return p.Type }
	Key() K
	// Value return a copy of itself
	//
	// case:
	// func (p *Proto) Value() Proto { return *p }
	Value() T
	// Self return a pointer of itself
	//
	// case:
	// func (p *Proto) Self() *Proto { return p }
	Self() *T
}

type Resetter interface {
	Reset()
}

// New a proto engine instance
//
// args
//
// - newProto should be return a Proto instance
func New[T any, K comparable](newProto NewProtoFunc[T, K]) *Engine[T, K] {
	return &Engine[T, K]{
		codec:    &CodecJSON[T, K]{},
		handlers: make(map[K]HandlerFunc[T]),
		newProto: newProto,
	}
}
