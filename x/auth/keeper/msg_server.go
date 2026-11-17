package keeper

import (
	"context"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	ak AccountKeeper
}

// NewMsgServerImpl returns an implementation of the x/auth MsgServer interface.
//
// Deprecated: register the service with RegisterMsgServer and run msgs from
// outside the module with msgservice.Dispatch. NewMsgServerImpl will be
// removed in a future release.
func NewMsgServerImpl(ak AccountKeeper) types.MsgServer {
	return &msgServer{
		ak: ak,
	}
}

// RegisterMsgServer registers the x/auth Msg service on r. The msg server
// is unexported: code outside the module runs a msg through the
// MsgServiceRouter with msgservice.Dispatch.
func RegisterMsgServer(r grpc.ServiceRegistrar, ak AccountKeeper) {
	types.RegisterMsgServer(r, NewMsgServerImpl(ak))
}

func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := sdk.ValidateAuthority(ctx, ms.ak.authority, msg.Authority); err != nil {
		return nil, err
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := ms.ak.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
