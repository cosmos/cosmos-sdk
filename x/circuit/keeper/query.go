package keeper

import (
	"bytes"
	"context"

	"cosmossdk.io/store/prefix"
	"cosmossdk.io/x/circuit/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var _ types.QueryServer = QueryServer{}

type QueryServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the circuit MsgServer interface
// for the provided Keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &QueryServer{keeper: keeper}
}

// Account returns account permissions.
func (qs QueryServer) Account(c context.Context, req *types.QueryAccountRequest) (*types.AccountResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(c)

	add, err := qs.keeper.addressCodec.StringToBytes(req.Address)
	if err != nil {
		return nil, err
	}

	perms, err := qs.keeper.Permissions.Get(sdkCtx, add)
	if err != nil {
		return nil, err
	}

	return &types.AccountResponse{Permission: &perms}, nil
}

// Account returns account permissions.
func (qs QueryServer) Accounts(ctx context.Context, req *types.QueryAccountsRequest) (*types.AccountsResponse, error) {
	// Iterate over accounts and perform the callback

	var accounts []*types.GenesisAccountPermissions
	store := runtime.KVStoreAdapter(qs.keeper.storeService.OpenKVStore(ctx))
	accountsStore := prefix.NewStore(store, types.AccountPermissionPrefix)

	pageRes, err := query.Paginate(accountsStore, req.Pagination, func(key, value []byte) error {
		perm := &types.Permissions{}
		if err := proto.Unmarshal(value, perm); err != nil {
			return err
		}

		// remove key suffix
		address, err := qs.keeper.addressCodec.BytesToString(bytes.TrimRight(key, "\x00"))
		if err != nil {
			return err
		}

		accounts = append(accounts, &types.GenesisAccountPermissions{
			Address:     address,
			Permissions: perm,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.AccountsResponse{Accounts: accounts, Pagination: pageRes}, nil
}

// DisabledList returns a list of disabled message urls
func (qs QueryServer) DisabledList(ctx context.Context, req *types.QueryDisabledListRequest) (*types.DisabledListResponse, error) {
	// Iterate over disabled list and perform the callback
	var msgs []string
	qs.keeper.DisableList.Walk(ctx, nil, func(msgUrl string) (bool, error) {
		msgs = append(msgs, msgUrl)
		return false, nil
	})

	return &types.DisabledListResponse{DisabledList: msgs}, nil
}
