package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

var _ quarantine.MsgServer = Keeper{}

func (k Keeper) OptIn(goCtx context.Context, msg *quarantine.MsgOptIn) (*quarantine.MsgOptInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if err = k.SetOptIn(ctx, toAddr); err != nil {
		return nil, err
	}

	return &quarantine.MsgOptInResponse{}, nil
}

func (k Keeper) OptOut(goCtx context.Context, msg *quarantine.MsgOptOut) (*quarantine.MsgOptOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if err = k.SetOptOut(ctx, toAddr); err != nil {
		return nil, err
	}

	return &quarantine.MsgOptOutResponse{}, nil
}

func (k Keeper) Accept(goCtx context.Context, msg *quarantine.MsgAccept) (*quarantine.MsgAcceptResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}

	fromAddrs := make([]sdk.AccAddress, len(msg.FromAddresses))
	for i, addr := range msg.FromAddresses {
		fromAddrs[i], err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address[%d]: %v", i, err)
		}
	}

	var fundsReleased sdk.Coins
	fundsReleased, err = k.AcceptQuarantinedFunds(ctx, toAddr, fromAddrs...)
	if err != nil {
		return nil, err
	}

	if msg.Permanent {
		for _, fromAddr := range fromAddrs {
			k.SetAutoResponse(ctx, toAddr, fromAddr, quarantine.AUTO_RESPONSE_ACCEPT)
		}
	}

	return &quarantine.MsgAcceptResponse{FundsReleased: fundsReleased}, nil
}

func (k Keeper) Decline(goCtx context.Context, msg *quarantine.MsgDecline) (*quarantine.MsgDeclineResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}

	fromAddrs := make([]sdk.AccAddress, len(msg.FromAddresses))
	for i, addr := range msg.FromAddresses {
		fromAddrs[i], err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address[%d]: %v", i, err)
		}
	}

	k.DeclineQuarantinedFunds(ctx, toAddr, fromAddrs...)

	if msg.Permanent {
		for _, fromAddr := range fromAddrs {
			k.SetAutoResponse(ctx, toAddr, fromAddr, quarantine.AUTO_RESPONSE_DECLINE)
		}
	}

	return &quarantine.MsgDeclineResponse{}, nil
}

func (k Keeper) UpdateAutoResponses(goCtx context.Context, msg *quarantine.MsgUpdateAutoResponses) (*quarantine.MsgUpdateAutoResponsesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	toAddr, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %v", err)
	}

	for i, update := range msg.Updates {
		fromAddr, err := sdk.AccAddressFromBech32(update.FromAddress)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address[%d]: %v", i, err)
		}
		k.SetAutoResponse(ctx, toAddr, fromAddr, update.Response)
	}

	return &quarantine.MsgUpdateAutoResponsesResponse{}, nil
}
