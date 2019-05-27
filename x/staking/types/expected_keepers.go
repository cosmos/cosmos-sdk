package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
)

// expected coin keeper
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
	GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins
}

// expected fee collection keeper
type FeeCollectionKeeper interface {
	GetCollectedFees(ctx sdk.Context) sdk.Coins
}

// expected bank keeper
type BankKeeper interface {
	DelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
	UndelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
}

// AccountKeeper expected Account keeper
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(auth.Account) (stop bool))
}

// ValidatorInterface expected validator functions
type ValidatorInterface interface {
	IsJailed() bool                                             // whether the validator is jailed
	GetMoniker() string                                         // moniker of the validator
	GetStatus() BondStatus                                      // status of the validator
	GetOperator() sdk.ValAddress                                // operator address to receive/return validators coins
	GetConsPubKey() crypto.PubKey                               // validation consensus pubkey
	GetConsAddr() sdk.ConsAddress                               // validation consensus address
	GetTokens() sdk.Int                                         // validation tokens
	GetBondedTokens() sdk.Int                                   // validator bonded tokens
	GetTendermintPower() int64                                  // validation power in tendermint
	GetCommission() sdk.Dec                                     // validator commission rate
	GetMinSelfDelegation() sdk.Int                              // validator minimum self delegation
	GetDelegatorShares() sdk.Dec                                // total outstanding delegator shares
	TokensFromShares(sdk.Dec) sdk.Dec                           // token worth of provided delegator shares
	TokensFromSharesTruncated(sdk.Dec) sdk.Dec                  // token worth of provided delegator shares, truncated
	TokensFromSharesRoundUp(sdk.Dec) sdk.Dec                    // token worth of provided delegator shares, rounded up
	SharesFromTokens(amt sdk.Int) (sdk.Dec, sdk.Error)          // shares worth of delegator's bond
	SharesFromTokensTruncated(amt sdk.Int) (sdk.Dec, sdk.Error) // truncated shares worth of delegator's bond
}

// ValidatorSet expected properties for the set of all validators
type ValidatorSet interface {
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(sdk.Context,
		func(index int64, validator ValidatorInterface) (stop bool))

	// iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator ValidatorInterface) (stop bool))

	// iterate through the consensus validator set of the last block by operator address, execute func for each validator
	IterateLastValidators(sdk.Context,
		func(index int64, validator ValidatorInterface) (stop bool))

	Validator(sdk.Context, sdk.ValAddress) ValidatorInterface            // get a particular validator by operator address
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) ValidatorInterface // get a particular validator by consensus address
	TotalBondedTokens(sdk.Context) sdk.Int                               // total bonded tokens within the validator set
	TotalTokens(sdk.Context) sdk.Int                                     // total token supply

	// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
	Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec)
	Jail(sdk.Context, sdk.ConsAddress)   // jail a validator
	Unjail(sdk.Context, sdk.ConsAddress) // unjail a validator

	// Delegation allows for getting a particular delegation for a given validator
	// and delegator outside the scope of the staking module.
	Delegation(sdk.Context, sdk.AccAddress, sdk.ValAddress) DelegationInterface

	// MaxValidators returns the maximum amount of bonded validators
	MaxValidators(sdk.Context) uint16
}

// DelegationSet expected properties for the set of all delegations for a particular
type DelegationSet interface {
	GetValidatorSet() ValidatorSet // validator set for which delegation set is based upon

	// iterate through all delegations from one delegator by validator-AccAddress,
	//   execute func for each validator
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation DelegationInterface) (stop bool))
}

// DelegationInterface delegation bond for a delegated proof of stake system
type DelegationInterface interface {
	GetDelegatorAddr() sdk.AccAddress // delegator sdk.AccAddress for the bond
	GetValidatorAddr() sdk.ValAddress // validator operator address
	GetShares() sdk.Dec               // amount of validator's shares held in this delegation
}

//_______________________________________________________________________________
// Event Hooks
// These can be utilized to communicate between a staking keeper and another
// keeper which must take particular actions when validators/delegators change
// state. The second keeper must implement this interface, which then the
// staking keeper can call.

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
