package proto

import "sync/atomic"

const (
	B  uint64 = 1
	KB        = 1 << (10 * iota)
	MB
	GB
)

type Config struct {
	// maxPayloadSize Specifies the maximum packet size that can be handled, unit: byte
	maxPayloadSize atomic.Uint64

	// maxErrorCount closed after the count of client errors reaches or exceeds the maxErrorCount
	maxErrorCount atomic.Uint32
}

func (c *Config) SetMaxPayloadSize(val uint64) {
	c.maxPayloadSize.Swap(val)
}

func (c *Config) MaxPayloadSize() uint64 {
	return c.maxPayloadSize.Load()
}

func (c *Config) SetMaxErrorCount(val uint32) {
	c.maxErrorCount.Swap(val)
}

func (c *Config) MaxErrorCount() uint32 {
	return c.maxErrorCount.Load()
}
