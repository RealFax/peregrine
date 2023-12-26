package peregrine

import "time"

const (
	DefaultWorkerPoolExpiry      = time.Second * 10
	DefaultWorkerPoolNonBlocking = true

	DefaultConnTimeout = time.Second * 15
)
