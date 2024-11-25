package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/bank/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type handlers struct {
	*Keeper
}

// NewHandlers creates a new bank/v2 handlers
func NewHandlers(k *Keeper) handlers {
	return handlers{k}
}

// UpdateParams updates the parameters of the bank/v2 module.
func (h handlers) MsgUpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	authorityBytes, err := h.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(h.authority, authorityBytes) {
		expectedAuthority, err := h.addressCodec.BytesToString(h.authority)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("invalid authority; expected %s, got %s", expectedAuthority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := h.params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (h handlers) MsgSend(ctx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
	var (
		from, to []byte
		err      error
	)

	from, err = h.addressCodec.StringToBytes(msg.FromAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid from address: %s", err)
	}

	to, err = h.addressCodec.StringToBytes(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	// TODO: Check denom enable

	err = h.SendCoins(ctx, from, to, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendResponse{}, nil
}

func (h handlers) MsgMint(ctx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	authorityBytes, err := h.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(h.authority, authorityBytes) {
		expectedAuthority, err := h.addressCodec.BytesToString(h.authority)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("invalid authority; expected %s, got %s", expectedAuthority, msg.Authority)
	}

	to, err := h.addressCodec.StringToBytes(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}

	if !msg.Amount.IsValid() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	if !msg.Amount.IsAllPositive() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}

	// TODO: should mint to mint module then transfer?
	err = h.MintCoins(ctx, to, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgMintResponse{}, nil
}

// QueryParams queries the parameters of the bank/v2 module.
func (h handlers) QueryParams(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	params, err := h.params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// QueryBalance queries the parameters of the bank/v2 module.
func (h handlers) QueryBalance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	if req == nil {
		return nil, errors.New("empty request")
	}

	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	addr, err := h.addressCodec.StringToBytes(req.Address)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	balance := h.Keeper.GetBalance(ctx, addr, req.Denom)

	return &types.QueryBalanceResponse{Balance: &balance}, nil
}
