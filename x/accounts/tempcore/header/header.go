package header

import (
	"context"
	"time"
)

// Header defines a generic header interface.
type Header interface {
	GetHeight() uint64  // GetHeight returns the height of the block
	GetHash() []byte    // GetHash returns the hash of the block header
	GetTime() time.Time // GetTime returns the time of the block
	GetChainID() string // GetChainID returns the chain ID of the chain
	GetAppHash() []byte // GetAppHash used in the current block header
}

// Service defines the interface in which you can get header information,
// given the execution context.
type Service[H Header] interface {
	GetHeader(ctx context.Context) H
}
