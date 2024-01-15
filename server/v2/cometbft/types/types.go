package types

import (
	"context"

	corecomet "cosmossdk.io/core/comet"
	abci "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"
)

type VoteExtensionsHandler interface {
	ExtendVote(context.Context, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error)
	VerifyVoteExtension(context.Context, *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error)
}

// PeerFilter responds to p2p filtering queries from Tendermint
type PeerFilter func(info string) (*abci.QueryResponse, error)

type ConsensusInfo struct { // TODO: this is a mock, we need a proper proto.Message
	proto.Message
	corecomet.Info
}
