package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/staking module sentinel errors
//
// TODO: Many of these errors are redundant. They should be removed and replaced
// by sdkerrors.ErrInvalidRequest.
//
// REF: https://github.com/cosmos/cosmos-sdk/issues/5450
var (
	ErrEmptyValidatorAddr              = errorsmod.Register(ModuleName, 2, "empty validator address")
	ErrNoValidatorFound                = errorsmod.Register(ModuleName, 3, "validator does not exist")
	ErrValidatorOwnerExists            = errorsmod.Register(ModuleName, 4, "validator already exist for this operator address; must use new validator operator address")
	ErrValidatorPubKeyExists           = errorsmod.Register(ModuleName, 5, "validator already exist for this pubkey; must use new validator pubkey")
	ErrValidatorPubKeyTypeNotSupported = errorsmod.Register(ModuleName, 6, "validator pubkey type is not supported")
	ErrValidatorJailed                 = errorsmod.Register(ModuleName, 7, "validator for this address is currently jailed")
	ErrBadRemoveValidator              = errorsmod.Register(ModuleName, 8, "failed to remove validator")
	ErrCommissionNegative              = errorsmod.Register(ModuleName, 9, "commission must be positive")
	ErrCommissionHuge                  = errorsmod.Register(ModuleName, 10, "commission cannot be more than 100%")
	ErrCommissionGTMaxRate             = errorsmod.Register(ModuleName, 11, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime            = errorsmod.Register(ModuleName, 12, "commission cannot be changed more than once in 24h")
	ErrCommissionChangeRateNegative    = errorsmod.Register(ModuleName, 13, "commission change rate must be positive")
	ErrCommissionChangeRateGTMaxRate   = errorsmod.Register(ModuleName, 14, "commission change rate cannot be more than the max rate")
	ErrCommissionGTMaxChangeRate       = errorsmod.Register(ModuleName, 15, "commission cannot be changed more than max change rate")
	ErrSelfDelegationBelowMinimum      = errorsmod.Register(ModuleName, 16, "validator's self delegation must be greater than their minimum self delegation")
	ErrMinSelfDelegationDecreased      = errorsmod.Register(ModuleName, 17, "minimum self delegation cannot be decrease")
	ErrEmptyDelegatorAddr              = errorsmod.Register(ModuleName, 18, "empty delegator address")
	ErrNoDelegation                    = errorsmod.Register(ModuleName, 19, "no delegation for (address, validator) tuple")
	ErrBadDelegatorAddr                = errorsmod.Register(ModuleName, 20, "delegator does not exist with address")
	ErrNoDelegatorForAddress           = errorsmod.Register(ModuleName, 21, "delegator does not contain delegation")
	ErrInsufficientShares              = errorsmod.Register(ModuleName, 22, "insufficient delegation shares")
	ErrDelegationValidatorEmpty        = errorsmod.Register(ModuleName, 23, "cannot delegate to an empty validator")
	ErrNotEnoughDelegationShares       = errorsmod.Register(ModuleName, 24, "not enough delegation shares")
	ErrNotMature                       = errorsmod.Register(ModuleName, 25, "entry not mature")
	ErrNoUnbondingDelegation           = errorsmod.Register(ModuleName, 26, "no unbonding delegation found")
	ErrMaxUnbondingDelegationEntries   = errorsmod.Register(ModuleName, 27, "too many unbonding delegation entries for (delegator, validator) tuple")
	ErrNoRedelegation                  = errorsmod.Register(ModuleName, 28, "no redelegation found")
	ErrSelfRedelegation                = errorsmod.Register(ModuleName, 29, "cannot redelegate to the same validator")
	ErrTinyRedelegationAmount          = errorsmod.Register(ModuleName, 30, "too few tokens to redelegate (truncates to zero tokens)")
	ErrBadRedelegationDst              = errorsmod.Register(ModuleName, 31, "redelegation destination validator not found")
	ErrTransitiveRedelegation          = errorsmod.Register(ModuleName, 32, "redelegation to this validator already in progress; first redelegation to this validator must complete before next redelegation")
	ErrMaxRedelegationEntries          = errorsmod.Register(ModuleName, 33, "too many redelegation entries for (delegator, src-validator, dst-validator) tuple")
	ErrDelegatorShareExRateInvalid     = errorsmod.Register(ModuleName, 34, "cannot delegate to validators with invalid (zero) ex-rate")
	ErrBothShareMsgsGiven              = errorsmod.Register(ModuleName, 35, "both shares amount and shares percent provided")
	ErrNeitherShareMsgsGiven           = errorsmod.Register(ModuleName, 36, "neither shares amount nor shares percent provided")
	ErrInvalidHistoricalInfo           = errorsmod.Register(ModuleName, 37, "invalid historical info")
	ErrNoHistoricalInfo                = errorsmod.Register(ModuleName, 38, "no historical info found")
	ErrEmptyValidatorPubKey            = errorsmod.Register(ModuleName, 39, "empty validator public key")
	ErrCommissionLTMinRate             = errorsmod.Register(ModuleName, 40, "commission cannot be less than min rate")
	ErrUnbondingNotFound               = errorsmod.Register(ModuleName, 41, "unbonding operation not found")
	ErrUnbondingOnHoldRefCountNegative = errorsmod.Register(ModuleName, 42, "cannot un-hold unbonding operation that is not on hold")
)
