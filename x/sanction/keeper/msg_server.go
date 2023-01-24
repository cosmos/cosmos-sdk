package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/errors"
)

var _ sanction.MsgServer = Keeper{}

func (k Keeper) Sanction(goCtx context.Context, req *sanction.MsgSanction) (*sanction.MsgSanctionResponse, error) {
	if req.Authority != k.authority {
		return nil, gov.ErrInvalidSigner.Wrapf("expected %q got %q", k.authority, req.Authority)
	}

	toSanction, err := toAccAddrs(req.Addresses)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrap(err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err = k.SanctionAddresses(ctx, toSanction...)
	if err != nil {
		return nil, err
	}

	return &sanction.MsgSanctionResponse{}, nil
}

func (k Keeper) Unsanction(goCtx context.Context, req *sanction.MsgUnsanction) (*sanction.MsgUnsanctionResponse, error) {
	if req.Authority != k.authority {
		return nil, gov.ErrInvalidSigner.Wrapf("expected %q got %q", k.authority, req.Authority)
	}

	toUnsanction, err := toAccAddrs(req.Addresses)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrap(err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err = k.UnsanctionAddresses(ctx, toUnsanction...)
	if err != nil {
		return nil, err
	}

	return &sanction.MsgUnsanctionResponse{}, nil
}

func (k Keeper) UpdateParams(goCtx context.Context, req *sanction.MsgUpdateParams) (*sanction.MsgUpdateParamsResponse, error) {
	if req.Authority != k.authority {
		return nil, gov.ErrInvalidSigner.Wrapf("expected %q got %q", k.authority, req.Authority)
	}

	if req.Params != nil {
		err := req.Params.ValidateBasic()
		if err != nil {
			return nil, errors.ErrInvalidParams.Wrap(err.Error())
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.SetParams(ctx, req.Params)
	if err != nil {
		return nil, err
	}

	return &sanction.MsgUpdateParamsResponse{}, nil
}
