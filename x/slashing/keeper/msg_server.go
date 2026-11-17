package keeper

import (
	"context"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the slashing MsgServer interface
// for the provided Keeper.
//
// Deprecated: register the service with RegisterMsgServer and run msgs from
// outside the module with msgservice.Dispatch. NewMsgServerImpl will be
// removed in a future release.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterMsgServer registers the x/slashing Msg service on r. The msg server
// is unexported: code outside the module runs a msg through the
// MsgServiceRouter with msgservice.Dispatch.
func RegisterMsgServer(r grpc.ServiceRegistrar, k Keeper) {
	types.RegisterMsgServer(r, NewMsgServerImpl(k))
}

// UpdateParams implements MsgServer.UpdateParams method.
// It defines a method to update the x/slashing module parameters.
func (k msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := sdk.ValidateAuthority(ctx, k.authority, msg.Authority); err != nil {
		return nil, err
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// Unjail implements MsgServer.Unjail method.
// Validators must submit a transaction to unjail itself after
// having been jailed (and thus unbonded) for downtime
func (k msgServer) Unjail(goCtx context.Context, msg *types.MsgUnjail) (*types.MsgUnjailResponse, error) {
	valAddr, err := k.sk.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddr)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("validator input address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.Unjail(ctx, valAddr); err != nil {
		return nil, err
	}

	return &types.MsgUnjailResponse{}, nil
}
