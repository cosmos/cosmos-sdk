package keeper

import (
	"context"
	"errors"
	"sort"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServer(k AccountKeeper) types.QueryServer {
	return queryServer{k: k}
}

type queryServer struct{ k AccountKeeper }

func (s queryServer) AccountAddressByID(ctx context.Context, req *types.QueryAccountAddressByIDRequest) (*types.QueryAccountAddressByIDResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Id != 0 { // ignoring `0` case since it is default value.
		return nil, status.Error(codes.InvalidArgument, "requesting with id isn't supported, try to request using account-id")
	}

	accID := req.AccountId

	address, err := s.k.Accounts.Indexes.Number.MatchExact(ctx, accID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "account address not found with account number %d", accID)
	}

	return &types.QueryAccountAddressByIDResponse{AccountAddress: address.String()}, nil
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
func (s queryServer) Account(ctx context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "Address cannot be empty")
	}

	addr, err := s.k.addressCodec.StringToBytes(req.Address)
	if err != nil {
		return nil, err
	}
	account := s.k.GetAccount(ctx, addr)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", req.Address)
	}

	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &types.QueryAccountResponse{Account: any}, nil
}

// Params returns parameters of auth module
func (s queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := s.k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ModuleAccounts returns all the existing Module Accounts
func (s queryServer) ModuleAccounts(c context.Context, req *types.QueryModuleAccountsRequest) (*types.QueryModuleAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// For deterministic output, sort the permAddrs by module name.
	sortedPermAddrs := make([]string, 0, len(s.k.permAddrs))
	for moduleName := range s.k.permAddrs {
		sortedPermAddrs = append(sortedPermAddrs, moduleName)
	}
	sort.Strings(sortedPermAddrs)

	modAccounts := make([]*codectypes.Any, 0, len(s.k.permAddrs))

	for _, moduleName := range sortedPermAddrs {
		account := s.k.GetModuleAccount(ctx, moduleName)
		if account == nil {
			return nil, status.Errorf(codes.NotFound, "account %s not found", moduleName)
		}
		any, err := codectypes.NewAnyWithValue(account)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
		modAccounts = append(modAccounts, any)
	}

	return &types.QueryModuleAccountsResponse{Accounts: modAccounts}, nil
}

// ModuleAccountByName returns module account by module name
func (s queryServer) ModuleAccountByName(c context.Context, req *types.QueryModuleAccountByNameRequest) (*types.QueryModuleAccountByNameResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(req.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "module name is empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	moduleName := req.Name

	account := s.k.GetModuleAccount(ctx, moduleName)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", moduleName)
	}
	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &types.QueryModuleAccountByNameResponse{Account: any}, nil
}

// Bech32Prefix returns the keeper internally stored bech32 prefix.
func (s queryServer) Bech32Prefix(ctx context.Context, req *types.Bech32PrefixRequest) (*types.Bech32PrefixResponse, error) {
	bech32Prefix, err := s.k.getBech32Prefix()
	if err != nil {
		return nil, err
	}

	if bech32Prefix == "" {
		return &types.Bech32PrefixResponse{Bech32Prefix: "bech32 is not used on this chain"}, nil
	}

	return &types.Bech32PrefixResponse{Bech32Prefix: bech32Prefix}, nil
}

// AddressBytesToString converts an address from bytes to string, using the
// keeper's bech32 prefix.
func (s queryServer) AddressBytesToString(ctx context.Context, req *types.AddressBytesToStringRequest) (*types.AddressBytesToStringResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.AddressBytes) == 0 {
		return nil, errors.New("empty address bytes is not allowed")
	}

	text, err := s.k.addressCodec.BytesToString(req.AddressBytes)
	if err != nil {
		return nil, err
	}

	return &types.AddressBytesToStringResponse{AddressString: text}, nil
}

// AddressStringToBytes converts an address from string to bytes, using the
// keeper's bech32 prefix.
func (s queryServer) AddressStringToBytes(ctx context.Context, req *types.AddressStringToBytesRequest) (*types.AddressStringToBytesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(strings.TrimSpace(req.AddressString)) == 0 {
		return nil, errors.New("empty address string is not allowed")
	}

	bz, err := s.k.addressCodec.StringToBytes(req.AddressString)
	if err != nil {
		return nil, err
	}

	return &types.AddressStringToBytesResponse{AddressBytes: bz}, nil
}

// AccountInfo implements the AccountInfo query.
func (s queryServer) AccountInfo(ctx context.Context, req *types.QueryAccountInfoRequest) (*types.QueryAccountInfoResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := s.k.addressCodec.StringToBytes(req.Address)
	if err != nil {
		return nil, err
	}

	account := s.k.GetAccount(ctx, addr)
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", req.Address)
	}

	// if there is no public key, avoid serializing the nil value
	pubKey := account.GetPubKey()
	var pkAny *codectypes.Any
	if pubKey != nil {
		pkAny, err = codectypes.NewAnyWithValue(account.GetPubKey())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%s", err.Error())
		}
	}

	return &types.QueryAccountInfoResponse{
		Info: &types.BaseAccount{
			Address:       req.Address,
			PubKey:        pkAny,
			AccountNumber: account.GetAccountNumber(),
			Sequence:      account.GetSequence(),
		},
	}, nil
}

// RekeyedAccounts returns the accounts whose current public key has the given
// natural address, in ascending address order.
func (s queryServer) RekeyedAccounts(ctx context.Context, req *types.QueryRekeyedAccountsRequest) (*types.QueryRekeyedAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	keyAddr, err := s.parseAddress(req.Address)
	if err != nil {
		return nil, err
	}

	addresses, pageRes, err := query.CollectionPaginate(ctx, s.k.RekeyIndex, req.Pagination,
		func(key collections.Pair[sdk.AccAddress, sdk.AccAddress], _ collections.NoValue) (string, error) {
			return s.k.addressCodec.BytesToString(key.K2())
		}, query.WithCollectionPaginationPairPrefix[sdk.AccAddress, sdk.AccAddress](keyAddr))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if addresses == nil {
		addresses = []string{}
	}

	return &types.QueryRekeyedAccountsResponse{Addresses: addresses, Pagination: pageRes}, nil
}

// PubKeyHistory returns the public key rotation history of an account, in the
// order the rotations happened.
func (s queryServer) PubKeyHistory(ctx context.Context, req *types.QueryPubKeyHistoryRequest) (*types.QueryPubKeyHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := s.parseAddress(req.Address)
	if err != nil {
		return nil, err
	}

	entries, pageRes, err := query.CollectionPaginate(ctx, s.k.PubKeyHistory, req.Pagination,
		func(_ collections.Pair[sdk.AccAddress, uint64], entry types.PubKeyHistoryEntry) (types.PubKeyHistoryEntry, error) {
			return entry, nil
		}, query.WithCollectionPaginationPairPrefix[sdk.AccAddress, uint64](addr))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if entries == nil {
		entries = []types.PubKeyHistoryEntry{}
	}

	return &types.QueryPubKeyHistoryResponse{Entries: entries, Pagination: pageRes}, nil
}

// parseAddress decodes a non-empty address string, returning an
// InvalidArgument status error on failure.
func (s queryServer) parseAddress(address string) (sdk.AccAddress, error) {
	if strings.TrimSpace(address) == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	bz, err := s.k.addressCodec.StringToBytes(address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address %q: %s", address, err)
	}

	return bz, nil
}
