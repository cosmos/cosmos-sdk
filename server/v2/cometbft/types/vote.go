package types

import (
	"context"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
)

// VoteExtensionsHandler defines how to implement vote extension handlers
type VoteExtensionsHandler interface {
	ExtendVote(context.Context, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error)
	VerifyVoteExtension(context.Context, *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error)
}
