package types

import (
	"fmt"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
)

// ABCICodeType - combined codetype / codespace
type ABCICodeType uint32

// CodeType - code identifier within codespace
type CodeType uint16

// CodespaceType - codespace identifier
type CodespaceType uint16

// get the abci code from the local code and codespace
func ToABCICode(code CodeType) ABCICodeType {
	if code == CodeOK {
		return ABCICodeOK
	}
	return ABCICodeType((uint32(CodespaceRoot) << 16) | uint32(code))
}

// SDK error codes
const (
	// Using same code with sdk/errors.go to reduce confusion

	// ABCI error codes
	ABCICodeOK ABCICodeType = 0

	// Base error codes
	CodeOK             CodeType = 0
	CodeInternal       CodeType = 1
	CodeTxDecode       CodeType = 2
	CodeUnknownRequest CodeType = 6

	CodespaceRoot CodespaceType = 1
)

func unknownCodeMsg(code CodeType) string {
	return fmt.Sprintf("unknown code %d", code)
}

// nolint: gocyclo
func CodeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInternal:
		return "internal error"
	case CodeTxDecode:
		return "tx parse error"
	case CodeUnknownRequest:
		return "unknown request"
	default:
		return unknownCodeMsg(code)
	}
}

// ErrInternal is for internal "err"s
func ErrInternal(msg string) Error {
	return newError(CodeInternal, msg)
}

// ErrTxDecode is for syntatically invalid query request
func ErrTxDecode(msg string) Error {
	return newError(CodeTxDecode, msg)
}

// ErrUnknownRequest is for semantically invalid query request
func ErrUnknownRequest(msg string) Error {
	return newError(CodeUnknownRequest, msg)
}

type cmnError = cmn.Error

// Query error type
type Error interface {
	cmnError

	QueryResult() abci.ResponseQuery
}

func newError(code CodeType, format string, args ...interface{}) Error {
	if format == "" {
		format = CodeToDefaultMsg(code)
	}

	return &queryError{
		code:     code,
		cmnError: cmn.NewError(format, args...),
	}
}

type queryError struct {
	code CodeType
	cmnError
}

func parseCmnError(err string) string {
	if idx := strings.Index(err, "{"); idx != -1 {
		err = err[idx+1 : len(err)-1]
	}
	return err
}

// Copied from types/errors.go
// Implements Error
func (err *queryError) ABCILog() string {
	cdc := codec.New()
	parsedErrMsg := parseCmnError(err.cmnError.Error())
	jsonErr := humanReadableError{
		Code:    err.code,
		Message: parsedErrMsg,
	}
	bz, er := cdc.MarshalJSON(jsonErr)
	if er != nil {
		panic(er)
	}
	stringifiedJSON := string(bz)
	return stringifiedJSON
}

// Implements Error
func (err *queryError) QueryResult() abci.ResponseQuery {
	return abci.ResponseQuery{
		Code: uint32(ToABCICode(err.code)),
		Log:  err.ABCILog(),
	}
}

// TODO: Do we need Codespace/ABCICode here?
type humanReadableError struct {
	Codespace CodespaceType `json:"codespace"`
	Code      CodeType      `json:"code"`
	ABCICode  ABCICodeType  `json:"abci_code"`
	Message   string        `json:"message"`
}
