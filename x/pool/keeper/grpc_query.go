package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/pool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

// CommunityPool queries the community pool coins
func (k Querier) CommunityPool(ctx context.Context, req *types.QueryCommunityPoolRequest) (*types.QueryCommunityPoolResponse, error) {
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAccount == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccount))
	}
	amount := k.bankKeeper.GetAllBalances(ctx, moduleAccount.GetAddress())
	decCoins := sdk.NewDecCoinsFromCoins(amount...)

	return &types.QueryCommunityPoolResponse{Pool: decCoins}, nil
}
