//nolint
package coin

import (
	rawerr "errors"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/errors"
)

var (
	errNoAccount         = rawerr.New("No such account")
	errInsufficientFunds = rawerr.New("Insufficient Funds")
	errNoInputs          = rawerr.New("No Input Coins")
	errNoOutputs         = rawerr.New("No Output Coins")
	errInvalidAddress    = rawerr.New("Invalid Address")
	errInvalidCoins      = rawerr.New("Invalid Coins")
	errInvalidSequence   = rawerr.New("Invalid Sequence")
)

var (
	invalidInput   = abci.CodeType_BaseInvalidInput
	invalidOutput  = abci.CodeType_BaseInvalidOutput
	unknownAddress = abci.CodeType_BaseUnknownAddress
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

func ErrNoAccount() errors.TMError {
	return errors.WithCode(errNoAccount, unknownAddress)
}

func IsNoAccountErr(err error) bool {
	return errors.IsSameError(errNoAccount, err)
}

func ErrInvalidAddress() errors.TMError {
	return errors.WithCode(errInvalidAddress, invalidInput)
}
func IsInvalidAddressErr(err error) bool {
	return errors.IsSameError(errInvalidAddress, err)
}

func ErrInvalidCoins() errors.TMError {
	return errors.WithCode(errInvalidCoins, invalidInput)
}
func IsInvalidCoinsErr(err error) bool {
	return errors.IsSameError(errInvalidCoins, err)
}

func ErrInvalidSequence() errors.TMError {
	return errors.WithCode(errInvalidSequence, invalidInput)
}
func IsInvalidSequenceErr(err error) bool {
	return errors.IsSameError(errInvalidSequence, err)
}

func ErrInsufficientFunds() errors.TMError {
	return errors.WithCode(errInsufficientFunds, invalidInput)
}
func IsInsufficientFundsErr(err error) bool {
	return errors.IsSameError(errInsufficientFunds, err)
}

func ErrNoInputs() errors.TMError {
	return errors.WithCode(errNoInputs, invalidInput)
}
func IsNoInputsErr(err error) bool {
	return errors.IsSameError(errNoInputs, err)
}

func ErrNoOutputs() errors.TMError {
	return errors.WithCode(errNoOutputs, invalidOutput)
}
func IsNoOutputsErr(err error) bool {
	return errors.IsSameError(errNoOutputs, err)
}
