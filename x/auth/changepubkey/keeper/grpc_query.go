package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// PubKeyHistory queries account pubkey history details based on address.
func (k Keeper) PubKeyHistory(goCtx context.Context, req *types.QueryPubKeyHistoryRequest) (*types.QueryPubKeyHistoryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "Address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	history := k.GetPubKeyHistory(ctx, addr)
	return &types.QueryPubKeyHistoryResponse{History: history}, nil
}

// PubKeyHistoricalEntry queries account pubkey historical entry based on address and time.
func (k Keeper) PubKeyHistoricalEntry(goCtx context.Context, req *types.QueryPubKeyHistoricalEntryRequest) (*types.QueryPubKeyHistoricalEntryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "Address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	entry := k.GetPubKeyHistoricalEntry(ctx, addr, req.Time)
	return &types.QueryPubKeyHistoricalEntryResponse{Entry: entry}, nil
}

// LastPubKeyHistoricalEntry queries account's last pubkey historical entry based on address.
func (k Keeper) LastPubKeyHistoricalEntry(goCtx context.Context, req *types.QueryLastPubKeyHistoricalEntryRequest) (*types.QueryLastPubKeyHistoricalEntryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "Address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	entry := k.GetLastPubKeyHistoricalEntry(ctx, addr)
	return &types.QueryLastPubKeyHistoricalEntryResponse{Entry: entry}, nil
}

// CurrentPubKeyEntry queries account's current pubkey entry based on address.
func (k Keeper) CurrentPubKeyEntry(goCtx context.Context, req *types.QueryCurrentPubKeyEntryRequest) (*types.QueryCurrentPubKeyEntryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "Address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	entry := k.GetCurrentPubKeyEntry(ctx, addr)
	return &types.QueryCurrentPubKeyEntryResponse{Entry: entry}, nil
}
