package types

import "cosmossdk.io/errors"

// x/staking module sentinel errors
var (
	ErrNoValidatorFound                = errors.Register(ModuleName, 103, "validator does not exist")
	ErrValidatorOwnerExists            = errors.Register(ModuleName, 104, "validator already exist for this operator address; must use new validator operator address")
	ErrValidatorPubKeyExists           = errors.Register(ModuleName, 105, "validator already exist for this pubkey; must use new validator pubkey")
	ErrValidatorPubKeyTypeNotSupported = errors.Register(ModuleName, 106, "validator pubkey type is not supported")
	ErrValidatorJailed                 = errors.Register(ModuleName, 107, "validator for this address is currently jailed")
	ErrCommissionNegative              = errors.Register(ModuleName, 109, "commission must be positive")
	ErrCommissionHuge                  = errors.Register(ModuleName, 1010, "commission cannot be more than 100%")
	ErrCommissionGTMaxRate             = errors.Register(ModuleName, 1011, "commission cannot be more than the max rate")
	ErrCommissionUpdateTime            = errors.Register(ModuleName, 1012, "commission cannot be changed more than once in 24h")
	ErrCommissionChangeRateNegative    = errors.Register(ModuleName, 1013, "commission change rate must be positive")
	ErrCommissionChangeRateGTMaxRate   = errors.Register(ModuleName, 1014, "commission change rate cannot be more than the max rate")
	ErrCommissionGTMaxChangeRate       = errors.Register(ModuleName, 1015, "commission cannot be changed more than max change rate")
	ErrSelfDelegationBelowMinimum      = errors.Register(ModuleName, 1016, "validator's self delegation must be greater than their minimum self delegation")
	ErrInsufficientShares              = errors.Register(ModuleName, 1022, "insufficient delegation shares")
	ErrDelegatorShareExRateInvalid     = errors.Register(ModuleName, 1034, "cannot delegate to validators with invalid (zero) ex-rate")
	ErrEmptyValidatorPubKey            = errors.Register(ModuleName, 1039, "empty validator public key")
	ErrCommissionLTMinRate             = errors.Register(ModuleName, 1040, "commission cannot be less than min rate")

	// consensus key errors
	ErrConsensusPubKeyLenInvalid = errors.Register(ModuleName, 1048, "consensus pubkey len is invalid")
)
