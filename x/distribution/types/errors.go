package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/distribution module sentinel errors
var (
	ErrEmptyDelegatorAddr      = sdkerrors.New(ModuleName, 2, "delegator address is empty")
	ErrEmptyWithdrawAddr       = sdkerrors.New(ModuleName, 3, "withdraw address is empty")
	ErrEmptyValidatorAddr      = sdkerrors.New(ModuleName, 4, "validator address is empty")
	ErrEmptyDelegationDistInfo = sdkerrors.New(ModuleName, 5, "no delegation distribution info")
	ErrNoValidatorDistInfo     = sdkerrors.New(ModuleName, 6, "no validator distribution info")
	ErrNoValidatorCommission   = sdkerrors.New(ModuleName, 7, "no validator commission to withdraw")
	ErrSetWithdrawAddrDisabled = sdkerrors.New(ModuleName, 8, "set withdraw address disabled")
	ErrBadDistribution         = sdkerrors.New(ModuleName, 9, "community pool does not have sufficient coins to distribute")
	ErrInvalidProposalAmount   = sdkerrors.New(ModuleName, 10, "invalid community pool spend proposal amount")
	ErrEmptyProposalRecipient  = sdkerrors.New(ModuleName, 11, "invalid community pool spend proposal recipient")
	ErrNoValidatorExists       = sdkerrors.New(ModuleName, 12, "validator does not exist")
	ErrNoDelegationExists      = sdkerrors.New(ModuleName, 13, "delegation does not exist")
)
