package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/slashing module sentinel errors
var (
	ErrNoValidatorForAddress        = errorsmod.Register(ModuleName, 2, "address is not associated with any known validator")
	ErrBadValidatorAddr             = errorsmod.Register(ModuleName, 3, "validator does not exist for that address")
	ErrValidatorJailed              = errorsmod.Register(ModuleName, 4, "validator still jailed; cannot be unjailed")
	ErrValidatorNotJailed           = errorsmod.Register(ModuleName, 5, "validator not jailed; cannot be unjailed")
	ErrMissingSelfDelegation        = errorsmod.Register(ModuleName, 6, "validator has no self-delegation; cannot be unjailed")
	ErrSelfDelegationTooLowToUnjail = errorsmod.Register(ModuleName, 7, "validator's self delegation less than minimum; cannot be unjailed")
	ErrNoSigningInfoFound           = errorsmod.Register(ModuleName, 8, "no validator signing info found")
)
