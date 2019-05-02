package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	// ModuleName is the name of the module
	ModuleName = "supply"

	// StoreKey is the default store key for supply
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the supply store.
	QuerierRoute = StoreKey
)

// Keeper defines the keeper of the supply store
type Keeper struct {
	cdc *codec.Codec
	ak  AccountKeeper
	dk  DistributionKeeper
	fck FeeCollectionKeeper
	sk  StakingKeeper
}

// NewKeeper creates a new supply Keeper instance
func NewKeeper(cdc *codec.Codec, ak AccountKeeper, dk DistributionKeeper, fck FeeCollectionKeeper, sk StakingKeeper) Keeper {
	return Keeper{cdc, ak, dk, fck, sk}
}

// AccountsSupply returns the balances, vesting and liquid supplies regarding accounts
func (k Keeper) AccountsSupply(ctx sdk.Context) (balance, vesting, liquid sdk.Coins) {
	k.ak.IterateAccounts(ctx, func(acc auth.Account) bool {
		balance = balance.Add(acc.GetCoins())
		liquid = liquid.Add(acc.SpendableCoins(ctx.BlockHeader().Time))

		vacc, isVestingAccount := acc.(auth.VestingAccount)
		if isVestingAccount {
			vesting = vesting.Add(vacc.GetVestingCoins(ctx.BlockHeader().Time))
		}
		return false
	})
	return
}

// EscrowedSupply returns the sum of all the tokens escrowed by the KVstore
func (k Keeper) EscrowedSupply(ctx sdk.Context) (escrowed sdk.Coins) {
	bondedSupply := sdk.NewCoins(sdk.NewCoin(k.sk.BondDenom(ctx), k.sk.TotalBondedTokens(ctx)))
	collectedFees := k.fck.GetCollectedFees(ctx)
	communityPool, remainingCommunityPool := k.dk.GetFeePoolCommunityCoins(ctx).TruncateDecimal()
	totalRewards, remainingRewards := k.dk.GetTotalRewards(ctx).TruncateDecimal()

	remaining, _ := remainingCommunityPool.Add(remainingRewards).TruncateDecimal()

	return bondedSupply.
		Add(collectedFees).
		Add(communityPool).
		Add(totalRewards).
		Add(remaining)
}

// TotalSupply returns the sum of account balances (free and vesting coins) and locked tokens on escrow
func (k Keeper) TotalSupply(ctx sdk.Context) (total sdk.Coins) {
	accounts, _, _ := k.AccountsSupply(ctx)
	escrowed := k.EscrowedSupply(ctx)

	return accounts.Add(escrowed)
}
