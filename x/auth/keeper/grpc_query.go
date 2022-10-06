package keeper

import (
	"context"
	"errors"
	"sort"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ types.QueryServer = accountKeeper{}

func (ak AccountKeeper) AccountAddressByID(c context.Context, req *types.QueryAccountAddressByIDRequest) (*types.QueryAccountAddressByIDResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Id < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid account number")
	}

	ctx := sdk.UnwrapSDKContext(c)
	address := ak.GetAccountAddressByID(ctx, uint64(req.GetId()))
	if len(address) == 0 {
		return nil, status.Errorf(codes.NotFound, "account address not found with account number %d", req.Id)
	}

	return &types.QueryAccountAddressByIDResponse{AccountAddress: address}, nil
}

func (ak AccountKeeper) Accounts(c context.Context, req *types.QueryAccountsRequest) (*types.QueryAccountsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Id != 0 { // ignoring `0` case since it is default value.
		return nil, status.Error(codes.InvalidArgument, "requesting with id isn't supported, try to request using account-id")
	}

	accID := req.AccountId

		accounts = append(accounts, any)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryAccountAddressByIDResponse{AccountAddress: addr}, nil
}

func (s queryServer) Accounts(ctx context.Context, req *types.QueryAccountsRequest) (*types.QueryAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	accounts, pageRes, err := query.CollectionPaginate(
		ctx,
		s.k.Accounts,
		req.Pagination,
		func(_ sdk.AccAddress, value sdk.AccountI) (*codectypes.Any, error) {
			return codectypes.NewAnyWithValue(value)
		},
	)

	return &types.QueryAccountsResponse{Accounts: accounts, Pagination: pageRes}, err
}

// Account returns account details based on address
func (ak accountKeeper) Account(c context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "Address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}
	account := s.k.GetAccount(ctx, addr)
	if account == nil {
		xAccount, err := s.getFromXAccounts(ctx, addr)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "account %s not found", req.Address)
		}
		return &types.QueryAccountResponse{Account: xAccount.Account}, nil
	}

	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAccountResponse{Account: any}, nil
}

// Params returns parameters of auth module
func (ak accountKeeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params := s.k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ModuleAccountByName returns module account by module name
func (ak AccountKeeper) ModuleAccountByName(c context.Context, req *types.QueryModuleAccountByNameRequest) (*types.QueryModuleAccountByNameResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(req.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "module name is empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	moduleName := req.Name

	account := ak.GetModuleAccount(ctx, moduleName)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", moduleName)
	}
	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryModuleAccountByNameResponse{Account: any}, nil
}
