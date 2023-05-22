package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/distribution module sentinel errors
var (
	ErrEmptyDelegatorAddr      = errorsmod.Register(ModuleName, 2, "delegator address is empty")
	ErrEmptyWithdrawAddr       = errorsmod.Register(ModuleName, 3, "withdraw address is empty")
	ErrEmptyValidatorAddr      = errorsmod.Register(ModuleName, 4, "validator address is empty")
	ErrEmptyDelegationDistInfo = errorsmod.Register(ModuleName, 5, "no delegation distribution info")
	ErrNoValidatorDistInfo     = errorsmod.Register(ModuleName, 6, "no validator distribution info")
	ErrNoValidatorCommission   = errorsmod.Register(ModuleName, 7, "no validator commission to withdraw")
	ErrSetWithdrawAddrDisabled = errorsmod.Register(ModuleName, 8, "set withdraw address disabled")
	ErrBadDistribution         = errorsmod.Register(ModuleName, 9, "community pool does not have sufficient coins to distribute")
	ErrInvalidProposalAmount   = errorsmod.Register(ModuleName, 10, "invalid community pool spend proposal amount")
	ErrEmptyProposalRecipient  = errorsmod.Register(ModuleName, 11, "invalid community pool spend proposal recipient")
	ErrNoValidatorExists       = errorsmod.Register(ModuleName, 12, "validator does not exist")
	ErrNoDelegationExists      = errorsmod.Register(ModuleName, 13, "delegation does not exist")
)
