package errors

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
)

// ABCICodeType - combined codetype / codespace
type ABCICodeType uint32

// CodeType - code identifier within codespace
type CodeType uint16

// CodespaceType - codespace identifier
type CodespaceType uint16

// SDK error codes
const (

	// ABCI error codes
	ABCICodeOK ABCICodeType = 0

	// Root error codes
	CodeOK             CodeType = 0
	CodeInternal       CodeType = 1
	CodeUnknownRequest CodeType = 2

	// CodespaceRoot is a codespace for error codes in this file only.
	// Notice that 0 is an "unset" codespace, which can be overridden with
	// Error.WithDefaultCodespace().
	CodespaceUndefined CodespaceType = 0
	CodespaceRoot      CodespaceType = 1

	// Maximum reservable codespace (2^16 - 1)
	MaximumCodespace CodespaceType = 65535
)

// Error type
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

// get the abci code from the local code and codespace
func ToABCICode(space CodespaceType, code CodeType) ABCICodeType {
	// TODO: Make Tendermint more aware of codespaces.
	if space == CodespaceRoot && code == CodeOK {
		return ABCICodeOK
	}
	return ABCICodeType((uint32(space) << 16) | uint32(code))
}

// IsOK - is everything okay?
func (code ABCICodeType) IsOK() bool {
	if code == ABCICodeOK {
		return true
	}
	return false
}

// NewError - create an error
func NewError(codespace CodespaceType, code CodeType, msg string) Error {
	return &sdkError{
		codespace: codespace,
		code:      code,
		err:       cmn.NewErrorWithT(code, msg),
	}
}

// Unknown Request Error on Root Codespace
func ErrUnknownRequest(msg string) Error {
	return NewError(CodespaceRoot, CodeUnknownRequest, msg)
}

type StdError struct {
	codespace CodespaceType
	code      CodeType
	err       cmn.Error
}

// Implements ABCIError.
func (err *SdkError) Error() string {
	return fmt.Sprintf("Error{%d:%d,%#v}", err.codespace, err.code, err.err)
}

// Implements ABCIError.
func (err *SdkError) ABCICode() ABCICodeType {
	return ToABCICode(err.codespace, err.code)
}

// Implements Error.
func (err *SdkError) Codespace() CodespaceType {
	return err.codespace
}

// Implements Error.
func (err *SdkError) Code() CodeType {
	return err.code
}

// Implements ABCIError.
func (err *StdError) ABCILog() string {
	return fmt.Sprintf(`=== ABCI Log ===
Codespace: %v
Code:      %v
ABCICode:  %v
Error:     %#v
=== /ABCI Log ===
`, err.codespace, err.code, err.ABCICode(), err.err)
}

// Add tracing information with msg.
func (err *StdError) Trace(msg string) Error {
	return &SdkError{
		codespace: err.codespace,
		code:      err.code,
		err:       err.err.Trace(msg),
	}
}

// Implements Error.
func (err *SdkError) WithDefaultCodespace(cs CodespaceType) Error {
	codespace := err.codespace
	if codespace == CodespaceUndefined {
		codespace = cs
	}
	return &SdkError{
		codespace: codespace,
		code:      err.code,
		err:       err.err,
	}
}

func (err *baseError) T() interface{} {
	return err.err.T()
}

func (err *SdkError) Result() Result {
	return Result{
		Code: err.ABCICode(),
		Log:  err.ABCILog(),
	}
}

// QueryResult allows us to return sdk.Error.QueryResult() in query responses
func (err *SdkError) QueryResult() abci.ResponseQuery {
	return abci.ResponseQuery{
		Code: uint32(err.ABCICode()),
		Log:  err.ABCILog(),
	}
}
