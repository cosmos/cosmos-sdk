package types

import (
	"context"
	"github.com/cosmos/gogoproto/proto"

	abci "github.com/cometbft/cometbft/abci/types"

	corecomet "cosmossdk.io/core/comet"
)

// VoteExtensionsHandler defines how to implement vote extension handlers
type VoteExtensionsHandler interface {
	ExtendVote(context.Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error)
}

type ConsensusInfo struct { // TODO: this is a mock, we need a proper proto.Message
	proto.Message
	corecomet.Info
}
