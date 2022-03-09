package blockinfo

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service is a type which retrieves basic block info from a context independent
// of any specific Tendermint core version. Modules which need a specific
// Tendermint header should use a different service and should expect to need
// to update whenever Tendermint makes any changes.
type Service interface {
	GetBlockInfo(ctx context.Context) BlockInfo
}

// BlockInfo represents basic block info independent of any specific Tendermint
// core version.
type BlockInfo interface {
	ChainID() string
	Height() int64
	Time() *timestamppb.Timestamp
	Hash() []byte
}
