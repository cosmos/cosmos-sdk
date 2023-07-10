package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/circuit/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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
func (qs QueryServer) Account(c context.Context, req *types.QueryAccountRequest) (*types.AccountResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(c)

	addr, err := qs.keeper.addressCodec.StringToBytes(req.Address)
	if err != nil {
		return nil, err
	}

	perms, err := qs.keeper.Permissions.Get(sdkCtx, addr)
	if err != nil {
		return nil, err
	}

	return &types.AccountResponse{Permission: &perms}, nil
}

// Account returns account permissions.
func (qs QueryServer) Accounts(ctx context.Context, req *types.QueryAccountsRequest) (*types.AccountsResponse, error) {
	var accounts []*types.GenesisAccountPermissions
	results, pageRes, err := query.CollectionPaginate[[]byte, types.Permissions](ctx, qs.keeper.Permissions, req.Pagination)
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		result := result
		address, err := qs.keeper.addressCodec.BytesToString(result.Key)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, &types.GenesisAccountPermissions{
			Address:     address,
			Permissions: &result.Value,
		})
	}

	return &types.AccountsResponse{Accounts: accounts, Pagination: pageRes}, nil
}

// DisabledList returns a list of disabled message urls
func (qs QueryServer) DisabledList(ctx context.Context, req *types.QueryDisabledListRequest) (*types.DisabledListResponse, error) {
	// Iterate over disabled list and perform the callback
	var msgs []string
	err := qs.keeper.DisableList.Walk(ctx, nil, func(msgUrl string) (bool, error) {
		msgs = append(msgs, msgUrl)
		return false, nil
	})
	if err != nil && !errorsmod.IsOf(err, collections.ErrInvalidIterator) {
		return nil, err
	}

	return &types.DisabledListResponse{DisabledList: msgs}, nil
}
