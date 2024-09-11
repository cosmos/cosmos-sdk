package testutil

import (
	"context"

	bankkeeper "cosmossdk.io/x/bank/v2/keeper"
	"cosmossdk.io/x/bank/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FundAccount is a utility function that funds an account by minting and
// sending the coins to the address. This should be used for testing purposes
// only!
func FundAccount(ctx context.Context, bankKeeper bankkeeper.Keeper, authKeeper types.AuthKeeper, addr []byte, amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, types.MintModuleName, amounts); err != nil {
		return err
	}
	mintAddr := authKeeper.GetModuleAddress(types.MintModuleName)
	return bankKeeper.SendCoins(ctx, mintAddr, addr, amounts)
}
