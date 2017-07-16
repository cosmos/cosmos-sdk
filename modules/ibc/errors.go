package ibc

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/errors"
)

// nolint
var (
	errChainNotRegistered  = fmt.Errorf("Chain not registered")
	errChainAlreadyExists  = fmt.Errorf("Chain already exists")
	errNeedsIBCPermission  = fmt.Errorf("Needs app-permission to send IBC")
	errCannotSetPermission = fmt.Errorf("Requesting invalid permission on IBC")
	// errNotMember        = fmt.Errorf("Not a member")
	// errInsufficientSigs = fmt.Errorf("Not enough signatures")
	// errNoMembers        = fmt.Errorf("No members specified")
	// errTooManyMembers   = fmt.Errorf("Too many members specified")
	// errNotEnoughMembers = fmt.Errorf("Not enough members specified")

	IBCCodeChainNotRegistered  = abci.CodeType(1001)
	IBCCodeChainAlreadyExists  = abci.CodeType(1002)
	IBCCodePacketAlreadyExists = abci.CodeType(1003)
	IBCCodeUnknownHeight       = abci.CodeType(1004)
	IBCCodeInvalidCommit       = abci.CodeType(1005)
	IBCCodeInvalidProof        = abci.CodeType(1006)
	IBCCodeInvalidCall         = abci.CodeType(1007)
)

func ErrNotRegistered(chainID string) error {
	return errors.WithMessage(chainID, errChainNotRegistered, IBCCodeChainNotRegistered)
}
func IsNotRegisteredErr(err error) bool {
	return errors.IsSameError(errChainNotRegistered, err)
}

func ErrAlreadyRegistered(chainID string) error {
	return errors.WithMessage(chainID, errChainAlreadyExists, IBCCodeChainAlreadyExists)
}
func IsAlreadyRegistetedErr(err error) bool {
	return errors.IsSameError(errChainAlreadyExists, err)
}

func ErrNeedsIBCPermission() error {
	return errors.WithCode(errNeedsIBCPermission, IBCCodeInvalidCall)
}
func IsNeedsIBCPermissionErr(err error) bool {
	return errors.IsSameError(errNeedsIBCPermission, err)
}

func ErrCannotSetPermission() error {
	return errors.WithCode(errCannotSetPermission, IBCCodeInvalidCall)
}
func IsCannotSetPermissionErr(err error) bool {
	return errors.IsSameError(errCannotSetPermission, err)
}
