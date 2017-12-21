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
)

const (
	CodeInvalidInput   uint32 = 101
	CodeInvalidOutput  uint32 = 102
	CodeUnknownAddress uint32 = 103
	CodeUnknownRequest uint32 = errors.CodeUnknownRequest
)

func ErrNoAccount() errors.ABCIError {
	return errors.WithCode(errNoAccount, CodeUnknownAddress)
}

func ErrInvalidAddress() errors.ABCIError {
	return errors.WithCode(errInvalidAddress, CodeInvalidInput)
}

func ErrInvalidCoins() errors.ABCIError {
	return errors.WithCode(errInvalidCoins, CodeInvalidInput)
}

func ErrInsufficientFunds() errors.ABCIError {
	return errors.WithCode(errInsufficientFunds, CodeInvalidInput)
}

func ErrInsufficientCredit() errors.ABCIError {
	return errors.WithCode(errInsufficientCredit, CodeInvalidInput)
}

func ErrNoInputs() errors.ABCIError {
	return errors.WithCode(errNoInputs, CodeInvalidInput)
}

func ErrNoOutputs() errors.ABCIError {
	return errors.WithCode(errNoOutputs, CodeInvalidOutput)
}
