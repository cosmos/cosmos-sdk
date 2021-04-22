package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"strings"
)

type Error = error

type EnvelopedErr struct {
	Err Error
}

type Coder interface {
	ABCICode() uint32
}

type Causer interface {
	Cause() error
}

type Codespacer interface {
	Codespace() string
}

func (e EnvelopedErr) ABCICode() uint32 {
	err := e.Err
	if err == nil {
		return errors.SuccessABCICode
	}
	for {
		if c, ok := err.(Coder); ok {
			return c.ABCICode()
		}

		if c, ok := err.(Causer); ok {
			err = c.Cause()
		} else {
			return 1
		}
	}
}

func (e EnvelopedErr) Codespace() string {
	err := e.Err
	if err == nil {
		return ""
	}

	for {
		if c, ok := err.(Codespacer); ok {
			return c.Codespace()
		}

		if c, ok := err.(Causer); ok {
			err = c.Cause()
		} else {
			return errors.UndefinedCodespace
		}
	}
}

const (
	CodeOK          uint32 = 0
	CodeInternal    uint32 = 23
	CodeGasOverflow uint32 = 22
)

var (
	// Base error codes

	CodeTxDecode          uint32 = errors.ErrTxDecode.ABCICode()
	CodeInvalidSequence   uint32 = errors.ErrInvalidSequence.ABCICode()
	CodeUnauthorized      uint32 = errors.ErrUnauthorized.ABCICode()
	CodeInsufficientFunds uint32 = errors.ErrInsufficientFunds.ABCICode()
	CodeUnknownRequest    uint32 = errors.ErrUnknownRequest.ABCICode()
	CodeInvalidAddress    uint32 = errors.ErrInvalidAddress.ABCICode()
	CodeInvalidPubKey     uint32 = errors.ErrInvalidPubKey.ABCICode()
	CodeUnknownAddress    uint32 = errors.ErrUnknownAddress.ABCICode()
	CodeInsufficientCoins uint32 = errors.ErrInsufficientFunds.ABCICode()
	CodeInvalidCoins      uint32 = errors.ErrInvalidCoins.ABCICode()
	CodeOutOfGas          uint32 = errors.ErrOutOfGas.ABCICode()
	CodeMemoTooLarge      uint32 = errors.ErrMemoTooLarge.ABCICode()
	CodeInsufficientFee   uint32 = errors.ErrInsufficientFee.ABCICode()
	CodeTooManySignatures uint32 = errors.ErrTooManySignatures.ABCICode()
	CodeNoSignatures      uint32 = errors.ErrNoSignatures.ABCICode()
)

var errInternal = errors.Register(errors.UndefinedCodespace, CodeInternal, "internal")
var errGasOverflow = errors.Register(errors.RootCodespace, CodeGasOverflow, "gas overflow")

func (err EnvelopedErr) Result() (*Result, error) {
	return nil, err.Err
}

func (err EnvelopedErr) Error() string {
	return err.Err.Error()
}

// appends a message to the head of the given error
func AppendMsgToErr(msg string, err string) string {
	msgIdx := strings.Index(err, "message\":\"")
	if msgIdx != -1 {
		errMsg := err[msgIdx+len("message\":\"") : len(err)-2]
		errMsg = fmt.Sprintf("%s; %s", msg, errMsg)
		return fmt.Sprintf("%s%s%s",
			err[:msgIdx+len("message\":\"")],
			errMsg,
			err[len(err)-2:],
		)
	}
	return fmt.Sprintf("%s; %s", msg, err)
}

func ErrInternal(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errInternal, msg),
	}
}

func ErrTxDecode(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrTxDecode, msg),
	}
}
func ErrInvalidSequence(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrInvalidSequence, msg),
	}
}
func ErrUnauthorized(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrUnauthorized, msg),
	}
}
func ErrInsufficientFunds(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrInsufficientFunds, msg),
	}
}
func ErrUnknownRequest(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrUnknownRequest, msg),
	}
}
func ErrInvalidAddress(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrInvalidAddress, msg),
	}
}
func ErrUnknownAddress(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrUnknownAddress, msg),
	}
}
func ErrInvalidPubKey(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrInvalidPubKey, msg),
	}
}
func ErrInsufficientCoins(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrInsufficientFunds, msg),
	}
}
func ErrInvalidCoins(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrInvalidCoins, msg),
	}
}
func ErrOutOfGas(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrOutOfGas, msg),
	}
}
func ErrMemoTooLarge(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrMemoTooLarge, msg),
	}
}
func ErrInsufficientFee(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrInsufficientFee, msg),
	}
}
func ErrTooManySignatures(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrTooManySignatures, msg),
	}
}
func ErrNoSignatures(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errors.ErrNoSignatures, msg),
	}
}
func ErrGasOverflow(msg string) EnvelopedErr {
	return EnvelopedErr{
		errors.Wrapf(errGasOverflow, msg),
	}
}
