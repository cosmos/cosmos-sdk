package types

import (
	"fmt"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"

	abci "github.com/tendermint/tendermint/abci/types"
)

// CodeType - ABCI code identifier within codespace
type CodeType uint32

// CodespaceType - codespace identifier
type CodespaceType string

// IsOK - is everything okay?
func (code CodeType) IsOK() bool {
	if code == CodeOK {
		return true
	}
	return false
}

// SDK error codes
const (
	// Base error codes
	CodeOK             CodeType = 0
	CodeInternal       CodeType = 1
	CodeTxDecode       CodeType = 2
	CodeUnknownRequest CodeType = 6

	// CodespaceQuery is a codespace for error codes in this file only.
	CodespaceQuery CodespaceType = "query"
)

func unknownCodeMsg(code CodeType) string {
	return fmt.Sprintf("unknown code %d", code)
}

// NOTE: Don't stringer this, we'll put better messages in later.
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

//--------------------------------------------------------------------------------
// All errors are created via constructors so as to enable us to hijack them
// and inject stack traces if we really want to.

// nolint
func ErrInternal(msg string) Error {
	return newQueryError(CodeInternal, msg)
}
func ErrTxDecode(msg string) Error {
	return newQueryError(CodeTxDecode, msg)
}
func ErrUnknownRequest(msg string) Error {
	return newQueryError(CodeUnknownRequest, msg)
}

//----------------------------------------
// Error & queryError

type cmnError = cmn.Error

// sdk Error type
type Error interface {
	// Implements cmn.Error
	// Error() string
	// Stacktrace() cmn.Error
	// Trace(offset int, format string, args ...interface{}) cmn.Error
	// Data() interface{}
	cmnError

	Code() CodeType
	ABCILog() string
	QueryResult() abci.ResponseQuery
}

func newQueryError(code CodeType, format string, args ...interface{}) *queryError {
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

// Implements ABCIError.
func (err *queryError) Error() string {
	return fmt.Sprintf(`ERROR:
Codespace: %s
Code: %d
Message: %#v
`, CodespaceQuery, err.code, err.cmnError.Error())
}

// Implements Error.
func (err *queryError) Code() CodeType {
	return err.code
}

// Implements ABCIError.
func (err *queryError) ABCILog() string {
	cdc := codec.New()
	errMsg := err.cmnError.Error()
	jsonErr := humanReadableError{
		Codespace: CodespaceQuery,
		Code:      err.code,
		Message:   errMsg,
	}
	bz, er := cdc.MarshalJSON(jsonErr)
	if er != nil {
		panic(er)
	}
	stringifiedJSON := string(bz)
	return stringifiedJSON
}

// QueryResult allows us to return sdk.Error.QueryResult() in query responses
func (err *queryError) QueryResult() abci.ResponseQuery {
	return abci.ResponseQuery{
		Code:      uint32(err.Code()),
		Codespace: string(CodespaceQuery),
		Log:       err.ABCILog(),
	}
}

//----------------------------------------
// REST error utilities

// parses the error into an object-like struct for exporting
type humanReadableError struct {
	Codespace CodespaceType `json:"codespace"`
	Code      CodeType      `json:"code"`
	Message   string        `json:"message"`
}
