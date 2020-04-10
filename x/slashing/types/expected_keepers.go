package types // noalias

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// AccountKeeper expected account keeper
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(auth.Account) (stop bool))
}

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(sdk.Context,
		func(index int64, validator exported.ValidatorI) (stop bool))

	Validator(sdk.Context, sdk.ValAddress) exported.ValidatorI            // get a particular validator by operator address
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) exported.ValidatorI // get a particular validator by consensus address

	Jail(sdk.Context, sdk.ConsAddress)   // jail a validator
	Unjail(sdk.Context, sdk.ConsAddress) // unjail a validator

	// MaxValidators returns the maximum amount of bonded validators
	MaxValidators(sdk.Context) uint16

	// kick out the bonded validator in an epoch
	AppendAbandonedValidatorAddrs(ctx sdk.Context, ConsAddr sdk.ConsAddress)
}

// StakingHooks event hooks for staking validator object
type StakingHooks interface {
	AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)                           // Must be called when a validator is created
	AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) // Must be called when a validator is deleted

	AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) // Must be called when a validator is bonded
	/* required by okchain */
	AfterValidatorDestroyed(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress)	// Must be called when a validator is destroyed by tx
}
