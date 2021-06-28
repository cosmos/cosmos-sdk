package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) NFT(ctx context.Context, request *types.QueryNFTRequest) (*types.QueryNFTResponse, error) {
	panic("implement me")
}

func (k Keeper) NFTs(ctx context.Context, request *types.QueryNFTsRequest) (*types.QueryNFTsResponse, error) {
	panic("implement me")
}

func (k Keeper) NFTsOf(ctx context.Context, request *types.QueryNFTsOfRequest) (*types.QueryNFTsOfResponse, error) {
	panic("implement me")
}

func (k Keeper) Supply(ctx context.Context, request *types.QuerySupplyRequest) (*types.QuerySupplyResponse, error) {
	panic("implement me")
}

func (k Keeper) Balance(ctx context.Context, request *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	panic("implement me")
}

func (k Keeper) Type(ctx context.Context, request *types.QueryTypeRequest) (*types.QueryTypeResponse, error) {
	panic("implement me")
}

func (k Keeper) Types(ctx context.Context, request *types.QueryTypesRequest) (*types.QueryTypesResponse, error) {
	panic("implement me")
}
