package header

import (
	"context"
	"time"
)

type Service interface {
	GetHeaderInfo(context.Context) Info
}

type Info interface {
	GetHeight() int64      // GetHeight returns the height of the block
	GetHeaderHash() []byte // GetHeaderHash returns the hash of the block header
	GetTime() time.Time    // GetTime returns the time of the block
	GetChainID() string    // GetChainId returns the chain ID of the block
}
