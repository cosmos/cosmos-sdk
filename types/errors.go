package types

import (
	"fmt"

	cmn "github.com/tendermint/tmlibs/common"

	abci "github.com/tendermint/abci/types"
)

// ABCICodeType - combined codetype / codespace
type ABCICodeType uint32

// CodeType - code identifier within codespace
type CodeType uint16

// CodespaceType - codespace identifier
type CodespaceType uint16

// IsOK - is everything okay?
func (code ABCICodeType) IsOK() bool {
	if code == ABCICodeOK {
		return true
	}
	return false
}

// get the abci code from the local code and codespace
func ToABCICode(space CodespaceType, code CodeType) ABCICodeType {
	// TODO: Make Tendermint more aware of codespaces.
	if space == CodespaceRoot && code == CodeOK {
		return ABCICodeOK
	}
	return ABCICodeType((uint32(space) << 16) | uint32(code))
}

// SDK error codes
const (
	// ABCI error codes
	ABCICodeOK ABCICodeType = 0

	// Base error codes
	CodeOK                CodeType = 0
	CodeInternal          CodeType = 1
	CodeTxDecode          CodeType = 2
	CodeInvalidSequence   CodeType = 3
	CodeUnauthorized      CodeType = 4
	CodeInsufficientFunds CodeType = 5
	CodeUnknownRequest    CodeType = 6
	CodeInvalidAddress    CodeType = 7
	CodeInvalidPubKey     CodeType = 8
	CodeUnknownAddress    CodeType = 9
	CodeInsufficientCoins CodeType = 10
	CodeInvalidCoins      CodeType = 11
	CodeOutOfGas          CodeType = 12

	// CodespaceRoot is a codespace for error codes in this file only.
	// Notice that 0 is an "unset" codespace, which can be overridden with
	// Error.WithDefaultCodespace().
	CodespaceUndefined CodespaceType = 0
	CodespaceRoot      CodespaceType = 1

	// Maximum reservable codespace (2^16 - 1)
	MaximumCodespace CodespaceType = 65535
)

// NOTE: Don't stringer this, we'll put better messages in later.
func CodeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInternal:
		return "internal error"
	case CodeTxDecode:
		return "tx parse error"
	case CodeInvalidSequence:
		return "invalid sequence"
	case CodeUnauthorized:
		return "unauthorized"
	case CodeInsufficientFunds:
		return "insufficent funds"
	case CodeUnknownRequest:
		return "unknown request"
	case CodeInvalidAddress:
		return "invalid address"
	case CodeInvalidPubKey:
		return "invalid pubkey"
	case CodeUnknownAddress:
		return "unknown address"
	case CodeInsufficientCoins:
		return "insufficient coins"
	case CodeInvalidCoins:
		return "invalid coins"
	case CodeOutOfGas:
		return "out of gas"
	default:
		return fmt.Sprintf("unknown code %d", code)
	}
}

//--------------------------------------------------------------------------------
// All errors are created via constructors so as to enable us to hijack them
// and inject stack traces if we really want to.

// nolint
func ErrInternal(msg string) Error {
	return newErrorWithRootCodespace(CodeInternal, msg)
}
func ErrTxDecode(msg string) Error {
	return newErrorWithRootCodespace(CodeTxDecode, msg)
}
func ErrInvalidSequence(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidSequence, msg)
}
func ErrUnauthorized(msg string) Error {
	return newErrorWithRootCodespace(CodeUnauthorized, msg)
}
func ErrInsufficientFunds(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientFunds, msg)
}
func ErrUnknownRequest(msg string) Error {
	return newErrorWithRootCodespace(CodeUnknownRequest, msg)
}
func ErrInvalidAddress(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidAddress, msg)
}
func ErrUnknownAddress(msg string) Error {
	return newErrorWithRootCodespace(CodeUnknownAddress, msg)
}
func ErrInvalidPubKey(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidPubKey, msg)
}
func ErrInsufficientCoins(msg string) Error {
	return newErrorWithRootCodespace(CodeInsufficientCoins, msg)
}
func ErrInvalidCoins(msg string) Error {
	return newErrorWithRootCodespace(CodeInvalidCoins, msg)
}
func ErrOutOfGas(msg string) Error {
	return newErrorWithRootCodespace(CodeOutOfGas, msg)
}

//----------------------------------------
// Error & sdkError

// sdk Error type
type Error interface {
	Error() string
	Code() CodeType
	Codespace() CodespaceType
	ABCILog() string
	ABCICode() ABCICodeType
	WithDefaultCodespace(codespace CodespaceType) Error
	Trace(msg string) Error
	T() interface{}
	Result() Result
	QueryResult() abci.ResponseQuery
}

// NewError - create an error
func NewError(codespace CodespaceType, code CodeType, msg string) Error {
	return newError(codespace, code, msg)
}

func newErrorWithRootCodespace(code CodeType, msg string) *sdkError {
	return newError(CodespaceRoot, code, msg)
}

func newError(codespace CodespaceType, code CodeType, msg string) *sdkError {
	if msg == "" {
		msg = CodeToDefaultMsg(code)
	}
	return &sdkError{
		codespace: codespace,
		code:      code,
		err:       cmn.NewErrorWithT(code, msg),
	}
}

type sdkError struct {
	codespace CodespaceType
	code      CodeType
	err       cmn.Error
}

// Implements ABCIError.
func (err *sdkError) Error() string {
	return fmt.Sprintf("error{%d:%d,%#v}", err.codespace, err.code, err.err)
}

// Implements ABCIError.
func (err *sdkError) ABCICode() ABCICodeType {
	return ToABCICode(err.codespace, err.code)
}

// Implements Error.
func (err *sdkError) Codespace() CodespaceType {
	return err.codespace
}

// Implements Error.
func (err *sdkError) Code() CodeType {
	return err.code
}

// Implements ABCIError.
func (err *sdkError) ABCILog() string {
	return fmt.Sprintf(`=== ABCI Log ===
Codespace: %v
Code:      %v
ABCICode:  %v
Error:     %#v
=== /ABCI Log ===
`, err.codespace, err.code, err.ABCICode(), err.err)
}

// Add tracing information with msg.
func (err *sdkError) Trace(msg string) Error {
	return &sdkError{
		codespace: err.codespace,
		code:      err.code,
		err:       err.err.Trace(msg),
	}
}

// Implements Error.
func (err *sdkError) WithDefaultCodespace(cs CodespaceType) Error {
	codespace := err.codespace
	if codespace == CodespaceUndefined {
		codespace = cs
	}
	return &sdkError{
		codespace: codespace,
		code:      err.code,
		err:       err.err,
	}
}

func (err *sdkError) T() interface{} {
	return err.err.T()
}

func (err *sdkError) Result() Result {
	return Result{
		Code: err.ABCICode(),
		Log:  err.ABCILog(),
	}
}

// QueryResult allows us to return sdk.Error.QueryResult() in query responses
func (err *sdkError) QueryResult() abci.ResponseQuery {
	return abci.ResponseQuery{
		Code: uint32(err.ABCICode()),
		Log:  err.ABCILog(),
	}
}
