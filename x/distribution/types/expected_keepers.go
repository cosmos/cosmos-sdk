package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// expected staking keeper
type StakingKeeper interface {
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation sdk.Delegation) (stop bool))
	Delegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) sdk.Delegation
	Validator(ctx sdk.Context, valAddr sdk.ValAddress) sdk.Validator
	ValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) sdk.Validator
	GetLastTotalPower(ctx sdk.Context) sdk.Int
	GetLastValidatorPower(ctx sdk.Context, valAddr sdk.ValAddress) int64

	// used for invariants
	IterateValidators(ctx sdk.Context,
		fn func(index int64, validator sdk.Validator) (stop bool))
	GetAllSDKDelegations(ctx sdk.Context) []sdk.Delegation
}

// expected coin keeper
type BankKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
}

// expected fee collection keeper
type FeeCollectionKeeper interface {
	GetCollectedFees(ctx sdk.Context) sdk.Coins
	ClearCollectedFees(ctx sdk.Context)
}
