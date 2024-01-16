package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	"google.golang.org/protobuf/proto"

	corecomet "cosmossdk.io/core/comet"
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
