package keeper

import (
	"context"

	types "cosmossdk.io/x/accountlink/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServer(k Keeper) types.QueryServer {
	return &queryServer{k}
}

type queryServer struct {
	Keeper
}

func (q queryServer) AccountsQuery(ctx context.Context, request *types.AccountsQueryRequest) (*types.AccountsQueryResponse, error) {
	err := request.Validate()
	if err != nil {
		return nil, err
	}

	if accExists := q.authKeeper.HasAccount(ctx, sdk.AccAddress(request.Owner)); !accExists {
		return nil, types.ErrNonExistOwner
	}

	accData := q.GetAccountsByOwner(ctx, sdk.AccAddress(request.Owner), request.AccountType)
	return &types.AccountsQueryResponse{
		Addresses: accData.Addresses,
	}, nil
}
