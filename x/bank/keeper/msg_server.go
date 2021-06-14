package keeper

import (
	"context"
	"fmt"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type msgServer struct {
	BaseKeeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper BaseKeeper) types.MsgServer {
	return &msgServer{BaseKeeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) Send(goCtx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.IsSendEnabledCoins(ctx, msg.Amount...); err != nil {
		return nil, err
	}

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if k.BlockedAddr(to) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
	}

	err = k.SendCoins(ctx, from, to, msg.Amount)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, a := range msg.Amount {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "send"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgSendResponse{}, nil
}

func (k msgServer) MultiSend(goCtx context.Context, msg *types.MsgMultiSend) (*types.MsgMultiSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// NOTE: totalIn == totalOut should already have been checked
	for _, in := range msg.Inputs {
		if err := k.IsSendEnabledCoins(ctx, in.Coins...); err != nil {
			return nil, err
		}
	}

	for _, out := range msg.Outputs {
		accAddr, err := sdk.AccAddressFromBech32(out.Address)
		if err != nil {
			panic(err)
		}
		if k.BlockedAddr(accAddr) {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive transactions", out.Address)
		}
	}

	err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgMultiSendResponse{}, nil
}

func (k msgServer) Mint(goCtx context.Context, req *types.MsgMint) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	minter, err := sdk.AccAddressFromBech32(req.Minter)
	if err != nil {
		return nil, err
	}

	for _, coin := range req.Amount {
		denom := coin.Denom
		mgr := k.getDenomManager(denom)
		if mgr == nil {
			return nil, fmt.Errorf("no denom manager for denom %s", denom)
		}

		err = mgr.OnMint(ctx, minter, coin)
		if err != nil {
			return nil, err
		}

		err = k.addCoins(ctx, minter, []sdk.Coin{coin})
		if err != nil {
			return nil, err
		}

		supply := k.GetSupply(ctx, denom)
		supply = supply.Add(coin)
		k.setSupply(ctx, supply)

	}

	ctx.EventManager().EmitEvent(
		types.NewCoinMintEvent(minter, req.Amount),
	)

	return &types.MsgMintResponse{}, nil
}

func (k msgServer) Burn(goCtx context.Context, req *types.MsgBurn) (*types.MsgBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	burner, err := sdk.AccAddressFromBech32(req.Burner)
	if err != nil {
		return nil, err
	}

	for _, coin := range req.Amount {
		denom := coin.Denom
		mgr := k.getDenomManager(denom)
		if mgr == nil {
			return nil, fmt.Errorf("no denom manager for denom %s", denom)
		}

		err = mgr.OnBurn(ctx, burner, coin)
		if err != nil {
			return nil, err
		}

		err = k.subUnlockedCoins(ctx, burner, []sdk.Coin{coin})
		if err != nil {
			return nil, err
		}

		supply := k.GetSupply(ctx, denom)
		supply = supply.Sub(coin)
		k.setSupply(ctx, supply)
	}

	ctx.EventManager().EmitEvent(
		types.NewCoinBurnEvent(burner, req.Amount),
	)

	return &types.MsgBurnResponse{}, nil
}
