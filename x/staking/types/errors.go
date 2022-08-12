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
	ErrEmptyValidatorAddr               = sdkerrors.New(ModuleName, 2, "empty validator address")
	ErrNoValidatorFound                 = sdkerrors.New(ModuleName, 3, "validator does not exist")
	ErrValidatorOwnerExists             = sdkerrors.New(ModuleName, 4, "validator already exist for this operator address; must use new validator operator address")
	ErrValidatorPubKeyExists            = sdkerrors.New(ModuleName, 5, "validator already exist for this pubkey; must use new validator pubkey")
	ErrValidatorPubKeyTypeNotSupported  = sdkerrors.New(ModuleName, 6, "validator pubkey type is not supported")
	ErrValidatorJailed                  = sdkerrors.New(ModuleName, 7, "validator for this address is currently jailed")
	ErrBadRemoveValidator               = sdkerrors.New(ModuleName, 8, "failed to remove validator")
	ErrCommissionNegative               = sdkerrors.New(ModuleName, 9, "commission must be positive")
	ErrCommissionHuge                   = sdkerrors.New(ModuleName, 10, "commission cannot be more than 100%")
	ErrCommissionGTMaxRate              = sdkerrors.New(ModuleName, 11, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime             = sdkerrors.New(ModuleName, 12, "commission cannot be changed more than once in 24h")
	ErrCommissionChangeRateNegative     = sdkerrors.New(ModuleName, 13, "commission change rate must be positive")
	ErrCommissionChangeRateGTMaxRate    = sdkerrors.New(ModuleName, 14, "commission change rate cannot be more than the max rate")
	ErrCommissionGTMaxChangeRate        = sdkerrors.New(ModuleName, 15, "commission cannot be changed more than max change rate")
	ErrSelfDelegationBelowMinimum       = sdkerrors.New(ModuleName, 16, "validator's self delegation must be greater than their minimum self delegation")
	ErrMinSelfDelegationDecreased       = sdkerrors.New(ModuleName, 17, "minimum self delegation cannot be decrease")
	ErrEmptyDelegatorAddr               = sdkerrors.New(ModuleName, 18, "empty delegator address")
	ErrNoDelegation                     = sdkerrors.New(ModuleName, 19, "no delegation for (address, validator) tuple")
	ErrBadDelegatorAddr                 = sdkerrors.New(ModuleName, 20, "delegator does not exist with address")
	ErrNoDelegatorForAddress            = sdkerrors.New(ModuleName, 21, "delegator does not contain delegation")
	ErrInsufficientShares               = sdkerrors.New(ModuleName, 22, "insufficient delegation shares")
	ErrDelegationValidatorEmpty         = sdkerrors.New(ModuleName, 23, "cannot delegate to an empty validator")
	ErrNotEnoughDelegationShares        = sdkerrors.New(ModuleName, 24, "not enough delegation shares")
	ErrNotMature                        = sdkerrors.New(ModuleName, 25, "entry not mature")
	ErrNoUnbondingDelegation            = sdkerrors.New(ModuleName, 26, "no unbonding delegation found")
	ErrMaxUnbondingDelegationEntries    = sdkerrors.New(ModuleName, 27, "too many unbonding delegation entries for (delegator, validator) tuple")
	ErrNoRedelegation                   = sdkerrors.New(ModuleName, 28, "no redelegation found")
	ErrSelfRedelegation                 = sdkerrors.New(ModuleName, 29, "cannot redelegate to the same validator")
	ErrTinyRedelegationAmount           = sdkerrors.New(ModuleName, 30, "too few tokens to redelegate (truncates to zero tokens)")
	ErrBadRedelegationDst               = sdkerrors.New(ModuleName, 31, "redelegation destination validator not found")
	ErrTransitiveRedelegation           = sdkerrors.New(ModuleName, 32, "redelegation to this validator already in progress; first redelegation to this validator must complete before next redelegation")
	ErrMaxRedelegationEntries           = sdkerrors.New(ModuleName, 33, "too many redelegation entries for (delegator, src-validator, dst-validator) tuple")
	ErrDelegatorShareExRateInvalid      = sdkerrors.New(ModuleName, 34, "cannot delegate to validators with invalid (zero) ex-rate")
	ErrBothShareMsgsGiven               = sdkerrors.New(ModuleName, 35, "both shares amount and shares percent provided")
	ErrNeitherShareMsgsGiven            = sdkerrors.New(ModuleName, 36, "neither shares amount nor shares percent provided")
	ErrInvalidHistoricalInfo            = sdkerrors.New(ModuleName, 37, "invalid historical info")
	ErrNoHistoricalInfo                 = sdkerrors.New(ModuleName, 38, "no historical info found")
	ErrEmptyValidatorPubKey             = sdkerrors.New(ModuleName, 39, "empty validator public key")
	ErrNotEnoughBalance                 = sdkerrors.New(ModuleName, 40, "not enough balance")
	ErrTokenizeShareRecordNotExists     = sdkerrors.New(ModuleName, 41, "tokenize share record not exists")
	ErrTokenizeShareRecordAlreadyExists = sdkerrors.New(ModuleName, 42, "tokenize share record already exists")
	ErrNotTokenizeShareRecordOwner      = sdkerrors.New(ModuleName, 43, "not tokenize share record owner")
	ErrExceedingFreeVestingDelegations  = sdkerrors.New(ModuleName, 44, "trying to exceed vested free delegation for vesting account")
	ErrOnlyBondDenomAllowdForTokenize   = sdkerrors.New(ModuleName, 45, "only bond denom is allowed for tokenize")
)
