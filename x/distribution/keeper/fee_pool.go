package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DistributeFeePool distributes funds from the the community pool to a receiver address
func (k Keeper) DistributeFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) sdk.Error {
	communityPoolAcc, err := k.supplyKeeper.GetPoolAccountByName(ctx, CommunityPoolName)
	if err != nil {
		return err
	}

	if !communityPoolAcc.GetCoins().IsAllGTE(amount) {
		return types.ErrBadDistribution(k.codespace)
	}
	_, remainder := communityPoolAcc.GetDecCoins().TruncateDecimal()

	err = k.supplyKeeper.SendCoinsPoolToAccount(ctx, CommunityPoolName, receiveAddr, amount)
	if err != nil {
		return err
	}
	err2 := communityPoolAcc.SetDecCoins(communityPoolAcc.GetDecCoins().Add(remainder))
	if err2 != nil {
		return sdk.ErrInvalidCoins(err2.Error())
	}

	k.supplyKeeper.SetPoolAccount(ctx, communityPoolAcc)
	return nil
}
