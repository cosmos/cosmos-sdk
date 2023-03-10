package keeper

import (
	context "context"

	"strings"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
	"github.com/cosmos/gogoproto/proto"
)

var _ types.QueryServer = QueryServer{}

type QueryServer struct {
	keeper Keeper
}

// Account returns account permissions.
func (qs QueryServer) Account(c context.Context, req *types.QueryAccountRequest) (*types.AccountResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(c)

	add, err := qs.keeper.addressCodec.StringToBytes(req.Address)
	if err != nil {
		return nil, err
	}

	perms, err := qs.keeper.GetPermissions(sdkCtx, add)
	if err != nil {
		return nil, err
	}

	return &types.AccountResponse{Permission: perms}, nil
}

// Account returns account permissions.
func (qs QueryServer) Accounts(c context.Context, req *types.QueryAccountsRequest) (*types.AccountsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(c)
	// Iterate over accounts and perform the callback

	var accounts []*types.GenesisAccountPermissions
	store := sdkCtx.KVStore(qs.keeper.storekey)
	accountsStore := prefix.NewStore(store, types.AccountPermissionPrefix)

	pageRes, err := query.Paginate(accountsStore, req.Pagination, func(key, value []byte) error {
		perm := &types.Permissions{}
		if err := proto.Unmarshal(value, perm); err != nil {
			return err
		}

		trim := strings.TrimRight(string(key), "\x00")
		accounts = append(accounts, &types.GenesisAccountPermissions{
			Address:     string(trim),
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
func (qs QueryServer) DisabledList(c context.Context, req *types.QueryDisableListRequest) (*types.DisabledListResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(c)
	// Iterate over disabled list and perform the callback

	var msgs []string
	qs.keeper.IterateDisableLists(sdkCtx, func(address []byte, perm types.Permissions) (stop bool) {
		for _, url := range perm.LimitTypeUrls {
			msgs = append(msgs, url)
		}
		return false
	})

	return &types.DisabledListResponse{DisabledList: msgs}, nil
}
