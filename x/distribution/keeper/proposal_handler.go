package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func HandleCommunityPoolSpendProposal(ctx sdk.Context, k Keeper, p types.CommunityPoolSpendProposal) sdk.Error {
	feePool := k.GetFeePool(ctx)
	newPool, negative := feePool.CommunityPool.SafeSub(sdk.NewDecCoins(p.Amount))
	if negative {
		return types.ErrBadDistribution(k.codespace)
	}
	feePool.CommunityPool = newPool
	k.SetFeePool(ctx, feePool)
	_, err := k.bankKeeper.AddCoins(ctx, p.Recipient, p.Amount)
	if err != nil {
		return err
	}
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("Spent %s coins from the community pool to recipient %s", p.Amount, p.Recipient))
	return nil
}
