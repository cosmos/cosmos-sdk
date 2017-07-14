package ibc

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/errors"
)

// nolint
var (
	errChainNotRegistered = fmt.Errorf("Chain not registered")
	errChainAlreadyExists = fmt.Errorf("Chain already exists")
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
)

func ErrNotRegistered(chainID string) error {
	return errors.WithMessage(chainID, errChainNotRegistered, IBCCodeChainNotRegistered)
}
func IsNotRegistetedErr(err error) bool {
	return errors.IsSameError(errChainNotRegistered, err)
}

func ErrAlreadyRegistered(chainID string) error {
	return errors.WithMessage(chainID, errChainAlreadyExists, IBCCodeChainAlreadyExists)
}
func IsAlreadyRegistetedErr(err error) bool {
	return errors.IsSameError(errChainAlreadyExists, err)
}
