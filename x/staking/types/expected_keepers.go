package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// DistributionKeeper expected distribution keeper (noalias)
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
	GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins
}

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(authtypes.AccountI) (stop bool))
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI // only used for simulation

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(sdk.Context, authtypes.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SetBalances(ctx sdk.Context, addr sdk.AccAddress, balances sdk.Coins) error
	LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	GetSupply(ctx sdk.Context) bankexported.SupplyI

	SendCoinsFromModuleToModule(ctx sdk.Context, senderPool, recipientPool string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

// ValidatorSet expected properties for the set of all validators (noalias)
type ValidatorSet interface {
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	// iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	// iterate through the consensus validator set of the last block by operator address, execute func for each validator
	IterateLastValidators(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	Validator(sdk.Context, sdk.ValAddress) stakingexported.ValidatorI            // get a particular validator by operator address
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingexported.ValidatorI // get a particular validator by consensus address
	TotalBondedTokens(sdk.Context) sdk.Int                                       // total bonded tokens within the validator set
	StakingTokenSupply(sdk.Context) sdk.Int                                      // total staking token supply

	// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
	Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec)
	Jail(sdk.Context, sdk.ConsAddress)   // jail a validator
	Unjail(sdk.Context, sdk.ConsAddress) // unjail a validator

	// Delegation allows for getting a particular delegation for a given validator
	// and delegator outside the scope of the staking module.
	Delegation(sdk.Context, sdk.AccAddress, sdk.ValAddress) stakingexported.DelegationI

	// MaxValidators returns the maximum amount of bonded validators
	MaxValidators(sdk.Context) uint32
}

// DelegationSet expected properties for the set of all delegations for a particular (noalias)
type DelegationSet interface {
	GetValidatorSet() ValidatorSet // validator set for which delegation set is based upon

	// iterate through all delegations from one delegator by validator-AccAddress,
	//   execute func for each validator
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation stakingexported.DelegationI) (stop bool))
}

//_______________________________________________________________________________
// Event Hooks
// These can be utilized to communicate between a staking keeper and another
// keeper which must take particular actions when validators/delegators change
// state. The second keeper must implement this interface, which then the
// staking keeper can call.

// StakingHooks event hooks for staking validator object (noalias)
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
