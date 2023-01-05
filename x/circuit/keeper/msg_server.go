package keeper

import (
	context "context"

	"github.com/cosmos/cosmos-sdk/x/circuit/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (srv msgServer) AuthorizeCircuitBreaker(ctx context.Context, msg *types.MsgAuthorizeCircuitBreaker) (*types.MsgAuthorizeCircuitBreakerResponse, error) {
	return nil, nil
}

// TripCircuitBreaker pauses processing of Msg's in the state machine.
func (srv msgServer) TripCircuitBreaker(ctx context.Context, msg *types.MsgTripCircuitBreaker) (*types.MsgTripCircuitBreakerResponse, error) {
	return nil, nil
}

// ResetCircuitBreaker resumes processing of Msg's in the state machine that
// have been been paused using TripCircuitBreaker.
func (srv msgServer) ResetCircuitBreaker(ctx context.Context, msg *types.MsgResetCircuitBreaker) (*types.MsgResetCircuitBreakerResponse, error) {
	return nil, nil
}
