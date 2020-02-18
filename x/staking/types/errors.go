package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/staking module sentinel errors
//
// TODO: Many of these errors are redundant. They should be removed and replaced
// by sdkerrors.ErrInvalidRequest.
//
// REF: https://github.com/cosmos/cosmos-sdk/issues/5450
var (
	ErrEmptyValidatorAddr              = sdkerrors.Register(ModuleName, 2, "empty validator address")
	ErrBadValidatorAddr                = sdkerrors.Register(ModuleName, 3, "validator address is invalid")
	ErrNoValidatorFound                = sdkerrors.Register(ModuleName, 4, "validator does not exist")
	ErrValidatorOwnerExists            = sdkerrors.Register(ModuleName, 5, "validator already exist for this operator address; must use new validator operator address")
	ErrValidatorPubKeyExists           = sdkerrors.Register(ModuleName, 6, "validator already exist for this pubkey; must use new validator pubkey")
	ErrValidatorPubKeyTypeNotSupported = sdkerrors.Register(ModuleName, 7, "validator pubkey type is not supported")
	ErrValidatorJailed                 = sdkerrors.Register(ModuleName, 8, "validator for this address is currently jailed")
	ErrBadRemoveValidator              = sdkerrors.Register(ModuleName, 9, "failed to remove validator")
	ErrCommissionNegative              = sdkerrors.Register(ModuleName, 10, "commission must be positive")
	ErrCommissionHuge                  = sdkerrors.Register(ModuleName, 11, "commission cannot be more than 100%")
	ErrCommissionGTMaxRate             = sdkerrors.Register(ModuleName, 12, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime            = sdkerrors.Register(ModuleName, 13, "commission cannot be changed more than once in 24h")
	ErrCommissionChangeRateNegative    = sdkerrors.Register(ModuleName, 14, "commission change rate must be positive")
	ErrCommissionChangeRateGTMaxRate   = sdkerrors.Register(ModuleName, 15, "commission change rate cannot be more than the max rate")
	ErrCommissionGTMaxChangeRate       = sdkerrors.Register(ModuleName, 16, "commission cannot be changed more than max change rate")
	ErrSelfDelegationBelowMinimum      = sdkerrors.Register(ModuleName, 17, "validator's self delegation must be greater than their minimum self delegation")
	ErrMinSelfDelegationInvalid        = sdkerrors.Register(ModuleName, 18, "minimum self delegation must be a positive integer")
	ErrMinSelfDelegationDecreased      = sdkerrors.Register(ModuleName, 19, "minimum self delegation cannot be decrease")
	ErrEmptyDelegatorAddr              = sdkerrors.Register(ModuleName, 20, "empty delegator address")
	ErrBadDenom                        = sdkerrors.Register(ModuleName, 21, "invalid coin denomination")
	ErrBadDelegationAddr               = sdkerrors.Register(ModuleName, 22, "invalid address for (address, validator) tuple")
	ErrBadDelegationAmount             = sdkerrors.Register(ModuleName, 23, "invalid delegation amount")
	ErrNoDelegation                    = sdkerrors.Register(ModuleName, 24, "no delegation for (address, validator) tuple")
	ErrBadDelegatorAddr                = sdkerrors.Register(ModuleName, 25, "delegator does not exist with address")
	ErrNoDelegatorForAddress           = sdkerrors.Register(ModuleName, 26, "delegator does not contain delegation")
	ErrInsufficientShares              = sdkerrors.Register(ModuleName, 27, "insufficient delegation shares")
	ErrDelegationValidatorEmpty        = sdkerrors.Register(ModuleName, 28, "cannot delegate to an empty validator")
	ErrNotEnoughDelegationShares       = sdkerrors.Register(ModuleName, 29, "not enough delegation shares")
	ErrBadSharesAmount                 = sdkerrors.Register(ModuleName, 30, "invalid shares amount")
	ErrBadSharesPercent                = sdkerrors.Register(ModuleName, 31, "Invalid shares percent")
	ErrNotMature                       = sdkerrors.Register(ModuleName, 32, "entry not mature")
	ErrNoUnbondingDelegation           = sdkerrors.Register(ModuleName, 33, "no unbonding delegation found")
	ErrMaxUnbondingDelegationEntries   = sdkerrors.Register(ModuleName, 34, "too many unbonding delegation entries for (delegator, validator) tuple")
	ErrBadRedelegationAddr             = sdkerrors.Register(ModuleName, 35, "invalid address for (address, src-validator, dst-validator) tuple")
	ErrNoRedelegation                  = sdkerrors.Register(ModuleName, 36, "no redelegation found")
	ErrSelfRedelegation                = sdkerrors.Register(ModuleName, 37, "cannot redelegate to the same validator")
	ErrTinyRedelegationAmount          = sdkerrors.Register(ModuleName, 38, "too few tokens to redelegate (truncates to zero tokens)")
	ErrBadRedelegationDst              = sdkerrors.Register(ModuleName, 39, "redelegation destination validator not found")
	ErrTransitiveRedelegation          = sdkerrors.Register(ModuleName, 40, "redelegation to this validator already in progress; first redelegation to this validator must complete before next redelegation")
	ErrMaxRedelegationEntries          = sdkerrors.Register(ModuleName, 41, "too many redelegation entries for (delegator, src-validator, dst-validator) tuple")
	ErrDelegatorShareExRateInvalid     = sdkerrors.Register(ModuleName, 42, "cannot delegate to validators with invalid (zero) ex-rate")
	ErrBothShareMsgsGiven              = sdkerrors.Register(ModuleName, 43, "both shares amount and shares percent provided")
	ErrNeitherShareMsgsGiven           = sdkerrors.Register(ModuleName, 44, "neither shares amount nor shares percent provided")
	ErrInvalidHistoricalInfo           = sdkerrors.Register(ModuleName, 45, "invalid historical info")
	ErrNoHistoricalInfo                = sdkerrors.Register(ModuleName, 46, "no historical info found")
	ErrEmptyValidatorPubKey            = sdkerrors.Register(ModuleName, 47, "empty validator public key")
)
