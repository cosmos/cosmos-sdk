package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/distribution module errors that reserve codes 300-399
var (
	ErrEmptyDelegatorAddr      = sdkerrors.Register(ModuleName, 300, "delegator address is empty")
	ErrEmptyWithdrawAddr       = sdkerrors.Register(ModuleName, 301, "withdraw address is empty")
	ErrEmptyValidatorAddr      = sdkerrors.Register(ModuleName, 302, "validator address is empty")
	ErrEmptyDelegationDistInfo = sdkerrors.Register(ModuleName, 303, "no delegation distribution info")
	ErrNoValidatorDistInfo     = sdkerrors.Register(ModuleName, 304, "no validator distribution info")
	ErrNoValidatorCommission   = sdkerrors.Register(ModuleName, 305, "no validator commission to withdraw")
	ErrSetWithdrawAddrDisabled = sdkerrors.Register(ModuleName, 306, "set withdraw address disabled")
	ErrBadDistribution         = sdkerrors.Register(ModuleName, 307, "community pool does not have sufficient coins to distribute")
	ErrInvalidProposalAmount   = sdkerrors.Register(ModuleName, 308, "invalid community pool spend proposal amount")
	ErrEmptyProposalRecipient  = sdkerrors.Register(ModuleName, 309, "invalid community pool spend proposal recipient")
	ErrNoValidatorExists       = sdkerrors.Register(ModuleName, 310, "validator does not exist")
	ErrNoDelegationExists      = sdkerrors.Register(ModuleName, 311, "delegation does not exist")
)
