package keeper

import (
	"context"

	types2 "github.com/cosmos/cosmos-sdk/contrib/x/circuit/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var _ types2.QueryServer = QueryServer{}

type QueryServer struct {
	keeper Keeper
}

// NewQueryServer returns an implementation of the circuit QueryServer interface
// for the provided Keeper.
func NewQueryServer(keeper Keeper) types2.QueryServer {
	return &QueryServer{keeper: keeper}
}

// Account returns account permissions.
func (qs QueryServer) Account(ctx context.Context, req *types2.QueryAccountRequest) (*types2.AccountResponse, error) {
	add, err := qs.keeper.addressCodec.StringToBytes(req.Address)
	if err != nil {
		return nil, err
	}

	perms, err := qs.keeper.Permissions.Get(ctx, add)
	if err != nil {
		return nil, err
	}

	return &types2.AccountResponse{Permission: &perms}, nil
}

// Accounts returns account permissions.
func (qs QueryServer) Accounts(ctx context.Context, req *types2.QueryAccountsRequest) (*types2.AccountsResponse, error) {
	results, pageRes, err := query.CollectionPaginate(
		ctx,
		qs.keeper.Permissions,
		req.Pagination,
		func(key []byte, value types2.Permissions) (*types2.GenesisAccountPermissions, error) {
			addrStr, err := qs.keeper.addressCodec.BytesToString(key)
			if err != nil {
				return nil, err
			}
			return &types2.GenesisAccountPermissions{
				Address:     addrStr,
				Permissions: &value,
			}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &types2.AccountsResponse{Accounts: results, Pagination: pageRes}, nil
}

// DisabledList returns a list of disabled message urls
func (qs QueryServer) DisabledList(ctx context.Context, req *types2.QueryDisabledListRequest) (*types2.DisabledListResponse, error) {
	// Iterate over disabled list and perform the callback
	var msgs []string
	err := qs.keeper.DisableList.Walk(ctx, nil, func(msgUrl string) (bool, error) {
		msgs = append(msgs, msgUrl)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &types2.DisabledListResponse{DisabledList: msgs}, nil
}
