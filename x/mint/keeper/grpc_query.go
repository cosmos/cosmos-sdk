package keeper

import (
	"context"

	"cosmossdk.io/x/mint/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Inflation returns minter.Inflation of the mint module.
func (q queryServer) Inflation(ctx context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	minter, err := q.k.Minter.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryInflationResponse{Inflation: minter.Inflation}, nil
}

// AnnualProvisions returns minter.AnnualProvisions of the mint module.
func (q queryServer) AnnualProvisions(ctx context.Context, _ *types.QueryAnnualProvisionsRequest) (*types.QueryAnnualProvisionsResponse, error) {
	minter, err := q.k.Minter.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryAnnualProvisionsResponse{AnnualProvisions: minter.AnnualProvisions}, nil
}

// GenesisTime returns minter.GenesisTime of the mint module.
func (q queryServer) GenesisTime(ctx context.Context, _ *types.QueryGenesisTimeRequest) (*types.QueryGenesisTimeResponse, error) {
	minter, err := q.k.Minter.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryGenesisTimeResponse{GenesisTime: minter.GenesisTime}, nil
}
