package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// AccountKeeper expected account keeper
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(auth.Account) (stop bool))
}

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(sdk.Context,
		func(index int64, validator staking.ValidatorInterface) (stop bool))

	// iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator staking.ValidatorInterface) (stop bool))

	// iterate through the consensus validator set of the last block by operator address, execute func for each validator
	IterateLastValidators(sdk.Context,
		func(index int64, validator staking.ValidatorInterface) (stop bool))

	Validator(sdk.Context, sdk.ValAddress) staking.ValidatorInterface            // get a particular validator by operator address
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) staking.ValidatorInterface // get a particular validator by consensus address
	TotalBondedTokens(sdk.Context) sdk.Int                                       // total bonded tokens within the validator set
	TotalTokens(sdk.Context) sdk.Int                                             // total token supply

	// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
	Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec)
	Jail(sdk.Context, sdk.ConsAddress)   // jail a validator
	Unjail(sdk.Context, sdk.ConsAddress) // unjail a validator

	// Delegation allows for getting a particular delegation for a given validator
	// and delegator outside the scope of the staking module.
	Delegation(sdk.Context, sdk.AccAddress, sdk.ValAddress) staking.DelegationInterface

	// MaxValidators returns the maximum amount of bonded validators
	MaxValidators(sdk.Context) uint16
}

// StakingHooks event hooks for staking validator object
type StakingHooks interface {
	AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)                           // Must be called when a validator is created
	BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress)                         // Must be called when a validator's state changes
	AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) // Must be called when a validator is deleted

	AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress)         // Must be called when a validator is bonded
	AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) // Must be called when a validator begins unbonding

	BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress)        // Must be called when a delegation is created
	BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) // Must be called when a delegation's shares are modified
	BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress)        // Must be called when a delegation is removed
	AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress)
	BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec)
}
