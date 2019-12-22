package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/slashing module errors that reserve codes 700-799
var (
	ErrNoValidatorForAddress        = sdkerrors.Register(ModuleName, 700, "address is not associated with any known validator")
	ErrBadValidatorAddr             = sdkerrors.Register(ModuleName, 701, "validator does not exist for that address")
	ErrValidatorJailed              = sdkerrors.Register(ModuleName, 702, "validator still jailed; cannot be unjailed")
	ErrValidatorNotJailed           = sdkerrors.Register(ModuleName, 703, "validator not jailed; cannot be unjailed")
	ErrMissingSelfDelegation        = sdkerrors.Register(ModuleName, 704, "validator has no self-delegation; cannot be unjailed")
	ErrSelfDelegationTooLowToUnjail = sdkerrors.Register(ModuleName, 705, "validator's self delegation less than minimum; cannot be unjailed")
	ErrNoSigningInfoFound           = sdkerrors.Register(ModuleName, 706, "no validator signing info found")
)
