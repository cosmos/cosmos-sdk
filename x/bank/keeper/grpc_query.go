package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ types.QueryServer = BaseKeeper{}

// Balance implements the Query/Balance gRPC method
func (k BaseKeeper) Balance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	balance := k.GetBalance(sdkCtx, address, req.Denom)

	return &types.QueryBalanceResponse{Balance: &balance}, nil
}

// AllBalances implements the Query/AllBalances gRPC method
func (k BaseKeeper) AllBalances(ctx context.Context, req *types.QueryAllBalancesRequest) (*types.QueryAllBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	balances := sdk.NewCoins()

	_, pageRes, err := query.CollectionFilteredPaginate(ctx, k.Balances, req.Pagination, func(key collections.Pair[sdk.AccAddress, string], value math.Int) (include bool, err error) {
		denom := key.K2()
		if req.ResolveDenom {
			if metadata, ok := k.GetDenomMetaData(sdkCtx, denom); ok {
				denom = metadata.Display
			}
		}
		balances = append(balances, sdk.NewCoin(denom, value))
		return false, nil // we don't include results because we're appending them here.
	}, query.WithCollectionPaginationPairPrefix[sdk.AccAddress, string](addr))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryAllBalancesResponse{Balances: balances, Pagination: pageRes}, nil
}

// SpendableBalances implements a gRPC query handler for retrieving an account's
// spendable balances.
func (k BaseKeeper) SpendableBalances(ctx context.Context, req *types.QuerySpendableBalancesRequest) (*types.QuerySpendableBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	balances := sdk.NewCoins()
	zeroAmt := math.ZeroInt()

	_, pageRes, err := query.CollectionFilteredPaginate(ctx, k.Balances, req.Pagination, func(key collections.Pair[sdk.AccAddress, string], _ math.Int) (include bool, err error) {
		balances = append(balances, sdk.NewCoin(key.K2(), zeroAmt))
		return false, nil // not including results as they're appended here
	}, query.WithCollectionPaginationPairPrefix[sdk.AccAddress, string](addr))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	result := sdk.NewCoins()
	spendable := k.SpendableCoins(sdkCtx, addr)

	for _, c := range balances {
		result = append(result, sdk.NewCoin(c.Denom, spendable.AmountOf(c.Denom)))
	}

	return &types.QuerySpendableBalancesResponse{Balances: result, Pagination: pageRes}, nil
}

// SpendableBalanceByDenom implements a gRPC query handler for retrieving an account's
// spendable balance for a specific denom.
func (k BaseKeeper) SpendableBalanceByDenom(ctx context.Context, req *types.QuerySpendableBalanceByDenomRequest) (*types.QuerySpendableBalanceByDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	spendable := k.SpendableCoin(sdkCtx, addr, req.Denom)

	return &types.QuerySpendableBalanceByDenomResponse{Balance: &spendable}, nil
}

// TotalSupply implements the Query/TotalSupply gRPC method
func (k BaseKeeper) TotalSupply(ctx context.Context, req *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	totalSupply, pageRes, err := k.GetPaginatedTotalSupply(sdkCtx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTotalSupplyResponse{Supply: totalSupply, Pagination: pageRes}, nil
}

// SupplyOf implements the Query/SupplyOf gRPC method
func (k BaseKeeper) SupplyOf(c context.Context, req *types.QuerySupplyOfRequest) (*types.QuerySupplyOfResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	supply := k.GetSupply(ctx, req.Denom)

	return &types.QuerySupplyOfResponse{Amount: sdk.NewCoin(req.Denom, supply.Amount)}, nil
}

// Params implements the gRPC service handler for querying x/bank parameters.
func (k BaseKeeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// DenomsMetadata implements Query/DenomsMetadata gRPC method.
func (k BaseKeeper) DenomsMetadata(c context.Context, req *types.QueryDenomsMetadataRequest) (*types.QueryDenomsMetadataResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomMetadataPrefix)

	metadatas := []types.Metadata{}
	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		var metadata types.Metadata
		k.cdc.MustUnmarshal(value, &metadata)

		metadatas = append(metadatas, metadata)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDenomsMetadataResponse{
		Metadatas:  metadatas,
		Pagination: pageRes,
	}, nil
}

// DenomMetadata implements Query/DenomMetadata gRPC method.
func (k BaseKeeper) DenomMetadata(c context.Context, req *types.QueryDenomMetadataRequest) (*types.QueryDenomMetadataResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	metadata, found := k.GetDenomMetaData(ctx, req.Denom)
	if !found {
		return nil, status.Errorf(codes.NotFound, "client metadata for denom %s", req.Denom)
	}

	return &types.QueryDenomMetadataResponse{
		Metadata: metadata,
	}, nil
}

func (k BaseKeeper) DenomOwners(
	goCtx context.Context,
	req *types.QueryDenomOwnersRequest,
) (*types.QueryDenomOwnersResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var denomOwners []*types.DenomOwner

	_, pageRes, err := query.CollectionFilteredPaginate(goCtx, k.Balances.Indexes.Denom, req.Pagination,
		func(key collections.Pair[string, sdk.AccAddress], value collections.NoValue) (include bool, err error) {
			amt, err := k.Balances.Get(goCtx, collections.Join(key.K2(), req.Denom))
			if err != nil {
				return false, err
			}
			denomOwners = append(denomOwners, &types.DenomOwner{
				Address: key.K2().String(),
				Balance: sdk.NewCoin(req.Denom, amt),
			})
			return false, nil
		},
		query.WithCollectionPaginationPairPrefix[string, sdk.AccAddress](req.Denom),
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryDenomOwnersResponse{DenomOwners: denomOwners, Pagination: pageRes}, nil
}

func (k BaseKeeper) SendEnabled(goCtx context.Context, req *types.QuerySendEnabledRequest) (*types.QuerySendEnabledResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &types.QuerySendEnabledResponse{}
	if len(req.Denoms) > 0 {
		for _, denom := range req.Denoms {
			if se, ok := k.getSendEnabled(ctx, denom); ok {
				resp.SendEnabled = append(resp.SendEnabled, types.NewSendEnabled(denom, se))
			}
		}
	} else {
		results, pageResp, err := query.CollectionPaginate[string, bool](ctx, k.BaseViewKeeper.SendEnabled, req.Pagination)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		for _, r := range results {
			resp.SendEnabled = append(resp.SendEnabled, &types.SendEnabled{
				Denom:   r.Key,
				Enabled: r.Value,
			})
		}
		resp.Pagination = pageResp
	}

	return resp, nil
}
