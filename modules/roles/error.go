//nolint
package roles

import (
	"fmt"

	abci "github.com/tendermint/abci/types"

	"github.com/tendermint/basecoin/errors"
)

var (
	errNoRole           = fmt.Errorf("No such role")
	errRoleExists       = fmt.Errorf("Role already exists")
	errNotMember        = fmt.Errorf("Not a member")
	errInsufficientSigs = fmt.Errorf("Not enough signatures")
	errNoMembers        = fmt.Errorf("No members specified")
	errTooManyMembers   = fmt.Errorf("Too many members specified")
	errNotEnoughMembers = fmt.Errorf("Not enough members specified")

	unauthorized = abci.CodeType_Unauthorized
)

// TODO: codegen?
// ex: err-gen NoRole,"No such role",CodeType_Unauthorized
func ErrNoRole() errors.TMError {
	return errors.WithCode(errNoRole, unauthorized)
}
func IsNoRoleErr(err error) bool {
	return errors.IsSameError(errNoRole, err)
}

func ErrRoleExists() errors.TMError {
	return errors.WithCode(errRoleExists, unauthorized)
}
func IsRoleExistsErr(err error) bool {
	return errors.IsSameError(errRoleExists, err)
}

func ErrNotMember() errors.TMError {
	return errors.WithCode(errNotMember, unauthorized)
}
func IsNotMemberErr(err error) bool {
	return errors.IsSameError(errNotMember, err)
}

func ErrInsufficientSigs() errors.TMError {
	return errors.WithCode(errInsufficientSigs, unauthorized)
}
func IsInsufficientSigsErr(err error) bool {
	return errors.IsSameError(errInsufficientSigs, err)
}

func ErrNoMembers() errors.TMError {
	return errors.WithCode(errNoMembers, unauthorized)
}
func IsNoMembersErr(err error) bool {
	return errors.IsSameError(errNoMembers, err)
}

func ErrTooManyMembers() errors.TMError {
	return errors.WithCode(errTooManyMembers, unauthorized)
}
func IsTooManyMembersErr(err error) bool {
	return errors.IsSameError(errTooManyMembers, err)
}

func ErrNotEnoughMembers() errors.TMError {
	return errors.WithCode(errNotEnoughMembers, unauthorized)
}
func IsNotEnoughMembersErr(err error) bool {
	return errors.IsSameError(errNotEnoughMembers, err)
}
