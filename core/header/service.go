package header

import (
	"context"
	"time"
)

type Service interface {
	GetHeaderInfo(context.Context) Info
}

type Info struct {
	Height  int64     // Height returns the height of the block
	Hash    []byte    // Hash returns the hash of the block header
	Time    time.Time // Time returns the time of the block
	ChainID string    // ChainId returns the chain ID of the block
}
