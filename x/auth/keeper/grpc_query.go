package keeper

import (
	"context"
	"errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/query"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ types.QueryServer = AccountKeeper{}

func (ak AccountKeeper) Accounts(c context.Context, req *types.QueryAccountsRequest) (*types.QueryAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(ak.key)
	accountsStore := prefix.NewStore(store, types.AddressStoreKeyPrefix)

	var accounts []*codectypes.Any
	pageRes, err := query.Paginate(accountsStore, req.Pagination, func(key, value []byte) error {
		account := ak.decodeAccount(value)
		any, err := codectypes.NewAnyWithValue(account)
		if err != nil {
			return err
		}

		accounts = append(accounts, any)
		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "paginate: %v", err)
	}

	return &types.QueryAccountsResponse{Accounts: accounts, Pagination: pageRes}, err
}

// Account returns account details based on address
func (ak AccountKeeper) Account(c context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
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
	account := ak.GetAccount(ctx, addr)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", req.Address)
	}

	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryAccountResponse{Account: any}, nil
}

// Params returns parameters of auth module
func (ak AccountKeeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := ak.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (ak AccountKeeper) Bech32Prefix(ctx context.Context, req *types.Bech32PrefixRequest) (*types.Bech32PrefixResponse, error) {
	bech32Prefix := ak.GetBech32Prefix()
	return &types.Bech32PrefixResponse{Bech32Prefix: bech32Prefix}, nil
}

func (ak AccountKeeper) Bech32FromAccAddr(ctx context.Context, req *types.Bech32FromAccAddrRequest) (*types.Bech32FromAccAddrResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(req.AccountAddr) == 0 {
		return nil, errors.New("empty bech32 string is not allowed")
	}

	bech32, err := bech32.ConvertAndEncode(ak.bech32Prefix, req.AccountAddr)
	if err != nil {
		return nil, err
	}

	return &types.Bech32FromAccAddrResponse{Bech32: bech32}, nil
}

func (ak AccountKeeper) AccAddrFromBech32(ctx context.Context, req *types.AccAddrFromBech32Request) (*types.AccAddrFromBech32Response, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(strings.TrimSpace(req.Bech32)) == 0 {
		return nil, errors.New("empty bech32 string is not allowed")
	}

	_, bz, err := bech32.DecodeAndConvert(req.Bech32)
	if err != nil {
		return nil, err
	}

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return nil, err
	}

	return &types.AccAddrFromBech32Response{AccountAddr: bz}, nil
}
