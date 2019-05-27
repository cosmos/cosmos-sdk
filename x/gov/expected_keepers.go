package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// expected bank keeper
type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	// TODO remove once governance doesn't require use of accounts
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SetSendEnabled(ctx sdk.Context, enabled bool)
}

// StakingKeeper expected staking keeper (Validator and Delegator sets)
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

	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation staking.DelegationInterface) (stop bool))
}
