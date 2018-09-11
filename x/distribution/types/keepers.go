package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// expected stake keeper
type StakeKeeper interface {
	IterateDelegations(ctx sdk.Context, delAddr sdk.AccAddress,
		fn func(index int64, del types.Delegation) (stop bool))
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress)
	GetValidator(ctx sdk.Context, valAddr sdk.AccAddress)
	GetPool(ctx sdk.Context)
}

// expected coin keeper
type BankKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
}

// from ante handler
type FeeCollectionKeeper interface {
	GetCollectedFees(ctx sdk.Context) sdk.Coins
	ClearCollectedFees(ctx sdk.Context)
}
