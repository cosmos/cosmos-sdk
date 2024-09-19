package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/bank/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the bank/v2 module genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	if err := k.params.Set(ctx, state.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	totalSupplyMap := sdk.NewMapCoins(sdk.Coins{})

	for _, balance := range state.Balances {
		addr := balance.Address
		bz, err := k.addressCodec.StringToBytes(addr)
		if err != nil {
			return err
		}

		for _, coin := range balance.Coins {
			err := k.balances.Set(ctx, collections.Join(bz, coin.Denom), coin.Amount)
			if err != nil {
				return err
			}
		}

		totalSupplyMap.Add(balance.Coins...)
	}
	totalSupply := totalSupplyMap.ToCoins()

	if !state.Supply.Empty() && !state.Supply.Equal(totalSupply) {
		return fmt.Errorf("genesis supply is incorrect, expected %v, got %v", state.Supply, totalSupply)
	}

	for _, supply := range totalSupply {
		k.setSupply(ctx, supply)
	}

	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return types.NewGenesisState(params), nil
}
