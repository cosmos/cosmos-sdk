package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetNextAccountNumber(ctx sdk.Context) uint64
}

// SupplyKeeper defines the supply Keeper for module accounts
type SupplyKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	GetModuleAccountByName(ctx sdk.Context, name string) supply.ModuleAccount
	SetModuleAccount(ctx sdk.Context, macc supply.ModuleAccount)

	SendCoinsModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SendCoinsAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
}

// StakingKeeper expected staking keeper (Validator and Delegator sets)
type StakingKeeper interface {
	// iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator exported.ValidatorI) (stop bool))

	TotalBondedTokens(sdk.Context) sdk.Int // total bonded tokens within the validator set

	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation exported.DelegationI) (stop bool))
}
