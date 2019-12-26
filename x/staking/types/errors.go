package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/staking module errors that reserve codes 800-899
//
// TODO: Many of these errors are redundant. They should be removed and replaced
// by sdkerrors.ErrInvalidRequest.
//
// REF: https://github.com/cosmos/cosmos-sdk/issues/5450
var (
	ErrEmptyValidatorAddr              = sdkerrors.Register(ModuleName, 800, "empty validator address")
	ErrBadValidatorAddr                = sdkerrors.Register(ModuleName, 801, "validator address is invalid")
	ErrNoValidatorFound                = sdkerrors.Register(ModuleName, 802, "validator does not exist")
	ErrValidatorOwnerExists            = sdkerrors.Register(ModuleName, 803, "validator already exist for this operator address; must use new validator operator address")
	ErrValidatorPubKeyExists           = sdkerrors.Register(ModuleName, 804, "validator already exist for this pubkey; must use new validator pubkey")
	ErrValidatorPubKeyTypeNotSupported = sdkerrors.Register(ModuleName, 805, "validator pubkey type is not supported")
	ErrValidatorJailed                 = sdkerrors.Register(ModuleName, 806, "validator for this address is currently jailed")
	ErrBadRemoveValidator              = sdkerrors.Register(ModuleName, 807, "failed to remove validator")
	ErrCommissionNegative              = sdkerrors.Register(ModuleName, 808, "commission must be positive")
	ErrCommissionHuge                  = sdkerrors.Register(ModuleName, 809, "commission cannot be more than 100%")
	ErrCommissionGTMaxRate             = sdkerrors.Register(ModuleName, 810, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime            = sdkerrors.Register(ModuleName, 811, "commission cannot be changed more than once in 24h")
	ErrCommissionChangeRateNegative    = sdkerrors.Register(ModuleName, 812, "commission change rate must be positive")
	ErrCommissionChangeRateGTMaxRate   = sdkerrors.Register(ModuleName, 813, "commission change rate cannot be more than the max rate")
	ErrCommissionGTMaxChangeRate       = sdkerrors.Register(ModuleName, 814, "commission cannot be changed more than max change rate")
	ErrSelfDelegationBelowMinimum      = sdkerrors.Register(ModuleName, 815, "validator's self delegation must be greater than their minimum self delegation")
	ErrMinSelfDelegationInvalid        = sdkerrors.Register(ModuleName, 816, "minimum self delegation must be a positive integer")
	ErrMinSelfDelegationDecreased      = sdkerrors.Register(ModuleName, 817, "minimum self delegation cannot be decrease")
	ErrEmptyDelegatorAddr              = sdkerrors.Register(ModuleName, 818, "empty delegator address")
	ErrBadDenom                        = sdkerrors.Register(ModuleName, 819, "invalid coin denomination")
	ErrBadDelegationAddr               = sdkerrors.Register(ModuleName, 820, "invalid address for (address, validator) tuple")
	ErrBadDelegationAmount             = sdkerrors.Register(ModuleName, 821, "invalid delegation amount")
	ErrNoDelegation                    = sdkerrors.Register(ModuleName, 822, "no delegation for (address, validator) tuple")
	ErrBadDelegatorAddr                = sdkerrors.Register(ModuleName, 823, "delegator does not exist with address")
	ErrNoDelegatorForAddress           = sdkerrors.Register(ModuleName, 824, "delegator does not contain delegation")
	ErrInsufficientShares              = sdkerrors.Register(ModuleName, 825, "insufficient delegation shares")
	ErrDelegationValidatorEmpty        = sdkerrors.Register(ModuleName, 826, "cannot delegate to an empty validator")
	ErrNotEnoughDelegationShares       = sdkerrors.Register(ModuleName, 827, "not enough delegation shares")
	ErrBadSharesAmount                 = sdkerrors.Register(ModuleName, 828, "invalid shares amount")
	ErrBadSharesPercent                = sdkerrors.Register(ModuleName, 829, "Invalid shares percent")
	ErrNotMature                       = sdkerrors.Register(ModuleName, 830, "entry not mature")
	ErrNoUnbondingDelegation           = sdkerrors.Register(ModuleName, 831, "no unbonding delegation found")
	ErrMaxUnbondingDelegationEntries   = sdkerrors.Register(ModuleName, 832, "too many unbonding delegation entries for (delegator, validator) tuple")
	ErrBadRedelegationAddr             = sdkerrors.Register(ModuleName, 833, "invalid address for (address, src-validator, dst-validator) tuple")
	ErrNoRedelegation                  = sdkerrors.Register(ModuleName, 834, "no redelegation found")
	ErrSelfRedelegation                = sdkerrors.Register(ModuleName, 835, "cannot redelegate to the same validator")
	ErrTinyRedelegationAmount          = sdkerrors.Register(ModuleName, 836, "too few tokens to redelegate (truncates to zero tokens)")
	ErrBadRedelegationDst              = sdkerrors.Register(ModuleName, 837, "redelegation destination validator not found")
	ErrTransitiveRedelegation          = sdkerrors.Register(ModuleName, 838, "redelegation to this validator already in progress; first redelegation to this validator must complete before next redelegation")
	ErrMaxRedelegationEntries          = sdkerrors.Register(ModuleName, 839, "too many redelegation entries for (delegator, src-validator, dst-validator) tuple")
	ErrDelegatorShareExRateInvalid     = sdkerrors.Register(ModuleName, 840, "cannot delegate to validators with invalid (zero) ex-rate")
	ErrBothShareMsgsGiven              = sdkerrors.Register(ModuleName, 841, "both shares amount and shares percent provided")
	ErrNeitherShareMsgsGiven           = sdkerrors.Register(ModuleName, 842, "neither shares amount nor shares percent provided")
	ErrInvalidHistoricalInfo           = sdkerrors.Register(ModuleName, 843, "invalid historical info")
	ErrNoHistoricalInfo                = sdkerrors.Register(ModuleName, 844, "no historical info found")
)
