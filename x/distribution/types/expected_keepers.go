package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

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

// SupplyKeeper defines the supply Keeper for pool accounts
type SupplyKeeper interface {
	GetPoolAccountByName(ctx sdk.Context, name string) supply.PoolAccount
	SetPoolAccount(ctx sdk.Context, pacc supply.PoolAccount)

	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsPoolToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SendCoinsAccountToPool(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
}
