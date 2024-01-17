package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	corecomet "cosmossdk.io/core/comet"
	"cosmossdk.io/server/v2/core/store"
)

type VoteExtensionsHandler interface {
	ExtendVote(context.Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error)
}

// PeerFilter responds to p2p filtering queries from Tendermint
type PeerFilter func(info string) (*abci.ResponseQuery, error)

type ConsensusInfo struct { // TODO: this is a mock, we need a proper proto.Message
	proto.Message
	corecomet.Info
}

type Store interface {
	store.Store
	Query(storeKey string, version uint64, key []byte, prove bool) (QueryResult, error)
}

type QueryResult interface {
	Key() []byte
	Value() []byte
	Version() uint64
	Proof() interface{} // CommitmentOp // TODO: use correct type
}
