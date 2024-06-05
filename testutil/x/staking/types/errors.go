package types

import "cosmossdk.io/errors"

// x/staking module sentinel errors
var (
	ErrNoValidatorFound                = errors.Register(ModuleName, 3, "validator does not exist")
	ErrValidatorOwnerExists            = errors.Register(ModuleName, 4, "validator already exist for this operator address; must use new validator operator address")
	ErrValidatorPubKeyExists           = errors.Register(ModuleName, 5, "validator already exist for this pubkey; must use new validator pubkey")
	ErrValidatorPubKeyTypeNotSupported = errors.Register(ModuleName, 6, "validator pubkey type is not supported")
	ErrValidatorJailed                 = errors.Register(ModuleName, 7, "validator for this address is currently jailed")
	ErrCommissionNegative              = errors.Register(ModuleName, 9, "commission must be positive")
	ErrCommissionHuge                  = errors.Register(ModuleName, 10, "commission cannot be more than 100%")
	ErrCommissionGTMaxRate             = errors.Register(ModuleName, 11, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime            = errors.Register(ModuleName, 12, "commission cannot be changed more than once in 24h")
	ErrCommissionChangeRateNegative    = errors.Register(ModuleName, 13, "commission change rate must be positive")
	ErrCommissionChangeRateGTMaxRate   = errors.Register(ModuleName, 14, "commission change rate cannot be more than the max rate")
	ErrCommissionGTMaxChangeRate       = errors.Register(ModuleName, 15, "commission cannot be changed more than max change rate")
	ErrSelfDelegationBelowMinimum      = errors.Register(ModuleName, 16, "validator's self delegation must be greater than their minimum self delegation")
	ErrInsufficientShares              = errors.Register(ModuleName, 22, "insufficient delegation shares")
	ErrDelegatorShareExRateInvalid     = errors.Register(ModuleName, 34, "cannot delegate to validators with invalid (zero) ex-rate")
	ErrEmptyValidatorPubKey            = errors.Register(ModuleName, 39, "empty validator public key")
	ErrCommissionLTMinRate             = errors.Register(ModuleName, 40, "commission cannot be less than min rate")

	// consensus key errors
	ErrConsensusPubKeyLenInvalid = errors.Register(ModuleName, 48, "consensus pubkey len is invalid")
)
