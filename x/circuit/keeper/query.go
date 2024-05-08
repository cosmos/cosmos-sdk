package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/x/circuit/types"
)

var _ types.QueryServer = QueryServer{}

type QueryServer struct {
	keeper Keeper
}

// NewQueryServer returns an implementation of the circuit QueryServer interface
// for the provided Keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &QueryServer{keeper: keeper}
}

// Account returns account permissions.
func (qs QueryServer) Account(ctx context.Context, req *types.QueryAccountRequest) (*types.AccountResponse, error) {
	add, err := qs.keeper.addressCodec.StringToBytes(req.Address)
	if err != nil {
		return nil, err
	}

	perms, err := qs.keeper.Permissions.Get(ctx, add)
	if err != nil {
		return nil, err
	}

	return &types.AccountResponse{Permission: &perms}, nil
}

// Account returns account permissions.
func (qs QueryServer) Accounts(ctx context.Context, req *types.QueryAccountsRequest) (*types.AccountsResponse, error) {
	return nil, errors.New("not implemented")
}

// DisabledList returns a list of disabled message urls
func (qs QueryServer) DisabledList(ctx context.Context, req *types.QueryDisabledListRequest) (*types.DisabledListResponse, error) {
	// Iterate over disabled list and perform the callback
	var msgs []string
	err := qs.keeper.DisableList.Walk(ctx, nil, func(msgUrl string) (bool, error) {
		msgs = append(msgs, msgUrl)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.DisabledListResponse{DisabledList: msgs}, nil
}
