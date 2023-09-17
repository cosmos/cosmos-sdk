package header

import (
	"context"
	"time"
)

// Service defines the interface in which you can get header information
type Service interface {
	GetHeaderInfo(context.Context) Info
}

// Info defines a struct that contains information about the header
type Info struct {
	Height  int64     // Height returns the height of the block
	Hash    []byte    // Hash returns the hash of the block header
	Time    time.Time // Time returns the time of the block
	ChainID string    // ChainId returns the chain ID of the block
	AppHash []byte    // AppHash used in the current block header
}

// Info implements the Header interface.
func (c Info) GetHeight() uint64  { return uint64(c.Height) } // check overflow/underflow
func (c Info) GetHash() []byte    { return c.Hash }
func (c Info) GetTime() time.Time { return c.Time }
func (c Info) GetChainID() string { return c.ChainID }
func (c Info) GetAppHash() []byte { return c.AppHash }
