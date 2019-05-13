package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func HandleCommunityPoolSpendProposal(ctx sdk.Context, k Keeper, p types.CommunityPoolSpendProposal) sdk.Error {
	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Sub(sdk.NewDecCoins(p.Amount))
	if feePool.CommunityPool.IsAnyNegative() {
		return types.ErrBadDistribution(k.codespace)
	}
	k.SetFeePool(ctx, feePool)
	_, err := k.bankKeeper.AddCoins(ctx, p.Recipient, p.Amount)
	if err != nil {
		return err
	}
	return nil
}
