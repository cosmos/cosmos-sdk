package keeper

import (
	"context"

	"cosmossdk.io/errors"
	types2 "github.com/cosmos/cosmos-sdk/contrib/x/crisis/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var _ types2.MsgServer = &Keeper{}

// VerifyInvariant implements MsgServer.VerifyInvariant method.
// It defines a method to verify a particular invariant.
func (k *Keeper) VerifyInvariant(goCtx context.Context, msg *types2.MsgVerifyInvariant) (*types2.MsgVerifyInvariantResponse, error) {
	if msg.Sender == "" {
		return nil, sdkerrors.ErrInvalidAddress.Wrap("empty address string is not allowed")
	}
	sender, err := k.addressCodec.StringToBytes(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := k.ConstantFee.Get(goCtx)
	if err != nil {
		return nil, err
	}
	constantFee := sdk.NewCoins(params)

	if err := k.SendCoinsFromAccountToFeeCollector(ctx, sender, constantFee); err != nil {
		return nil, err
	}

	// use a cached context to avoid gas costs during invariants
	cacheCtx, _ := ctx.CacheContext()

	found := false
	msgFullRoute := msg.FullInvariantRoute()

	var res string
	var stop bool
	for _, invarRoute := range k.Routes() {
		if invarRoute.FullRoute() == msgFullRoute {
			res, stop = invarRoute.Invar(cacheCtx)
			found = true

			break
		}
	}

	if !found {
		return nil, types2.ErrUnknownInvariant
	}

	if stop {
		// Currently, because the chain halts here, this transaction will never be included in the
		// blockchain thus the constant fee will have never been deducted. Thus no refund is required.

		// TODO replace with circuit breaker
		panic(res)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types2.EventTypeInvariant,
			sdk.NewAttribute(types2.AttributeKeyRoute, msg.InvariantRoute),
		),
	})

	return &types2.MsgVerifyInvariantResponse{}, nil
}

// UpdateParams implements MsgServer.UpdateParams method.
// It defines a method to update the x/crisis module parameters.
func (k *Keeper) UpdateParams(ctx context.Context, msg *types2.MsgUpdateParams) (*types2.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if !msg.ConstantFee.IsValid() {
		return nil, errors.Wrap(sdkerrors.ErrInvalidCoins, "invalid constant fee")
	}

	if msg.ConstantFee.IsNegative() {
		return nil, errors.Wrap(sdkerrors.ErrInvalidCoins, "negative constant fee")
	}

	if err := k.ConstantFee.Set(ctx, msg.ConstantFee); err != nil {
		return nil, err
	}

	return &types2.MsgUpdateParamsResponse{}, nil
}
