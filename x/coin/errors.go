//nolint
package coin

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/errors"
)

var (
	errNoAccount          = fmt.Errorf("No such account")
	errInsufficientFunds  = fmt.Errorf("Insufficient funds")
	errInsufficientCredit = fmt.Errorf("Insufficient credit")
	errNoInputs           = fmt.Errorf("No input coins")
	errNoOutputs          = fmt.Errorf("No output coins")
	errInvalidAddress     = fmt.Errorf("Invalid address")
	errInvalidCoins       = fmt.Errorf("Invalid coins")

	// TODO
	invalidInput   uint32 = 10
	invalidOutput  uint32 = 11
	unknownAddress uint32 = 12
	unknownRequest uint32 = errors.CodeUnknownRequest
)

// here are some generic handlers to grab classes of errors based on code
func IsInputErr(err error) bool {
	return errors.HasErrorCode(err, invalidInput)
}
func IsOutputErr(err error) bool {
	return errors.HasErrorCode(err, invalidOutput)
}
func IsAddressErr(err error) bool {
	return errors.HasErrorCode(err, unknownAddress)
}
func IsCoinErr(err error) bool {
	return err != nil && (IsInputErr(err) || IsOutputErr(err) || IsAddressErr(err))
}

func ErrNoAccount() errors.ABCIError {
	return errors.WithCode(errNoAccount, unknownAddress)
}

func IsNoAccountErr(err error) bool {
	return errors.IsSameError(errNoAccount, err)
}

func ErrInvalidAddress() errors.ABCIError {
	return errors.WithCode(errInvalidAddress, invalidInput)
}
func IsInvalidAddressErr(err error) bool {
	return errors.IsSameError(errInvalidAddress, err)
}

func ErrInvalidCoins() errors.ABCIError {
	return errors.WithCode(errInvalidCoins, invalidInput)
}
func IsInvalidCoinsErr(err error) bool {
	return errors.IsSameError(errInvalidCoins, err)
}

func ErrInsufficientFunds() errors.ABCIError {
	return errors.WithCode(errInsufficientFunds, invalidInput)
}
func IsInsufficientFundsErr(err error) bool {
	return errors.IsSameError(errInsufficientFunds, err)
}

func ErrInsufficientCredit() errors.ABCIError {
	return errors.WithCode(errInsufficientCredit, invalidInput)
}
func IsInsufficientCreditErr(err error) bool {
	return errors.IsSameError(errInsufficientCredit, err)
}

func ErrNoInputs() errors.ABCIError {
	return errors.WithCode(errNoInputs, invalidInput)
}
func IsNoInputsErr(err error) bool {
	return errors.IsSameError(errNoInputs, err)
}

func ErrNoOutputs() errors.ABCIError {
	return errors.WithCode(errNoOutputs, invalidOutput)
}
func IsNoOutputsErr(err error) bool {
	return errors.IsSameError(errNoOutputs, err)
}
