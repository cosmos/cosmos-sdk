package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

func (vk VestingKeeper) VestingAccounts(c context.Context, req *types.QueryVestingAccountsRequest) (*types.QueryVestingAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(vk.key)
	accountsStore := prefix.NewStore(store, types.VestingAccountStoreKeyPrefix)

	var accounts []*codectypes.Any
	pageRes, err := query.Paginate(accountsStore, req.Pagination, func(key, value []byte) error {
		addr := types.AddressFromVestingAccountKey(key)
		acct := vk.accountKeeper.GetAccount(ctx, addr)
		vestingAcct, ok := acct.(exported.VestingAccount)
		if !ok {
			return fmt.Errorf("account is not vesting account: %s", addr.String())
		}

		any, err := codectypes.NewAnyWithValue(vestingAcct)
		if err != nil {
			return err
		}
		accounts = append(accounts, any)
		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "paginate: %v", err)
	}

	return &types.QueryVestingAccountsResponse{Accounts: accounts, Pagination: pageRes}, err

}
