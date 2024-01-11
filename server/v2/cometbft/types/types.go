package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
)

type VoteExtensionsHandler interface {
	ExtendVote(context.Context, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error)
	VerifyVoteExtension(context.Context, *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error)
}

// PeerFilter responds to p2p filtering queries from Tendermint
type PeerFilter func(info string) (*abci.QueryResponse, error)
