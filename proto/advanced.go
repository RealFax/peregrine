package proto

func UseDefaultCodec[T any, K comparable]() Codec[T, K] {
	return &CodecJSON[T, K]{}
}
