package header

import (
	"context"
	"time"
)

type GetService[H Header] interface {
	GetHeader(context.Context) H
}

// Header defines a generic header interface.
type Header interface {
	GetHeight() uint64  // GetHeight returns the height of the block
	GetHash() []byte    // GetHash returns the hash of the block header
	GetTime() time.Time // GetTime returns the time of the block
	GetChainID() string // GetChainID returns the chain ID of the chain
	GetAppHash() []byte // GetAppHash used in the current block header
}
