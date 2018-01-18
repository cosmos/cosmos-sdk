package types

import (
	"fmt"
)

const (
	// ABCI Response Codes
	// Base SDK reserves 0 ~ 99.
	CodeInternalError       uint32 = 1
	CodeTxParseError               = 2
	CodeBadNonce                   = 3
	CodeUnauthorized               = 4
	CodeInsufficientFunds          = 5
	CodeUnknownRequest             = 6
	CodeUnrecognizedAddress        = 7
)

// NOTE: Don't stringer this, we'll put better messages in later.
func CodeToDefaultLog(code uint32) string {
	switch code {
	case CodeInternalError:
		return "Internal error"
	case CodeTxParseError:
		return "Tx parse error"
	case CodeBadNonce:
		return "Bad nonce"
	case CodeUnauthorized:
		return "Unauthorized"
	case CodeInsufficientFunds:
		return "Insufficent funds"
	case CodeUnknownRequest:
		return "Unknown request"
	case CodeUnrecognizedAddress:
		return "Unrecognized address"
	default:
		return fmt.Sprintf("Unknown code %d", code)
	}
}

//--------------------------------------------------------------------------------
// All errors are created via constructors so as to enable us to hijack them
// and inject stack traces if we really want to.

func InternalError(log string) Error {
	return newError(CodeInternalError, log)
}

func TxParseError(log string) Error {
	return newError(CodeTxParseError, log)
}

func BadNonce(log string) Error {
	return newError(CodeBadNonce, log)
}

func Unauthorized(log string) Error {
	return newError(CodeUnauthorized, log)
}

func InsufficientFunds(log string) Error {
	return newError(CodeInsufficientFunds, log)
}

func UnknownRequest(log string) Error {
	return newError(CodeUnknownRequest, log)
}

func UnrecognizedAddress(log string) Error {
	return newError(CodeUnrecognizedAddress, log)
}

//----------------------------------------
// Error & sdkError

type Error interface {
	Error() string
	ABCICode() uint32
	ABCILog() string
	Trace(msg string) Error
	TraceCause(cause error, msg string) Error
	Cause() error
	Result() Result
}

func NewError(code uint32, log string) Error {
	return newError(code, log)
}

type traceItem struct {
	msg      string
	filename string
	lineno   int
}

type sdkError struct {
	code  uint32
	log   string
	cause error
	trace []traceItem
}

func newError(code uint32, log string) *sdkError {
	// TODO capture stacktrace if ENV is set.
	if log == "" {
		log = CodeToDefaultLog(code)
	}
	return &sdkError{
		code:  code,
		log:   log,
		cause: nil,
		trace: nil,
	}
}

// Implements ABCIError.
func (err *sdkError) Error() string {
	return fmt.Sprintf("Error{%d:%s,%v,%v}", err.code, err.log, err.cause, len(err.trace))
}

// Implements ABCIError.
func (err *sdkError) ABCICode() uint32 {
	return err.code
}

// Implements ABCIError.
func (err *sdkError) ABCILog() string {
	return err.log
}

// Add tracing information to log with msg.
func (err *sdkError) Trace(msg string) Error {
	// Include file & line number & msg to log.
	// Do not include the whole stack trace.
	err.trace = append(err.trace, traceItem{
		filename: "todo", // TODO
		lineno:   -1,     // TODO
		msg:      msg,
	})
	return err
}

// Add tracing information to log with cause and msg.
func (err *sdkError) TraceCause(cause error, msg string) Error {
	err.cause = cause
	// Include file & line number & cause & msg to log.
	// Do not include the whole stack trace.
	err.trace = append(err.trace, traceItem{
		filename: "todo", // TODO
		lineno:   -1,     // TODO
		msg:      msg,
	})
	return err
}

func (err *sdkError) Cause() error {
	return err.cause
}

func (err *sdkError) Result() Result {
	return Result{
		Code: err.ABCICode(),
		Log:  err.ABCILog(),
	}
}
