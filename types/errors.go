package types

import (
	"fmt"
	"runtime"
)

const (
	// ABCI Response Codes
	// Base SDK reserves 0 ~ 99.
	CodeOK                  uint32 = 0
	CodeInternal                   = 1
	CodeTxParse                    = 2
	CodeBadNonce                   = 3
	CodeUnauthorized               = 4
	CodeInsufficientFunds          = 5
	CodeUnknownRequest             = 6
	CodeUnrecognizedAddress        = 7
	CodeInvalidSequence            = 8
)

// NOTE: Don't stringer this, we'll put better messages in later.
func CodeToDefaultMsg(code uint32) string {
	switch code {
	case CodeInternal:
		return "Internal error"
	case CodeTxParse:
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
	case CodeInvalidSequence:
		return "Invalid sequence"
	default:
		return fmt.Sprintf("Unknown code %d", code)
	}
}

//--------------------------------------------------------------------------------
// All errors are created via constructors so as to enable us to hijack them
// and inject stack traces if we really want to.

func ErrInternal(msg string) Error {
	return newError(CodeInternal, msg)
}

func ErrTxParse(msg string) Error {
	return newError(CodeTxParse, msg)
}

func ErrBadNonce(msg string) Error {
	return newError(CodeBadNonce, msg)
}

func ErrUnauthorized(msg string) Error {
	return newError(CodeUnauthorized, msg)
}

func ErrInsufficientFunds(msg string) Error {
	return newError(CodeInsufficientFunds, msg)
}

func ErrUnknownRequest(msg string) Error {
	return newError(CodeUnknownRequest, msg)
}

func ErrUnrecognizedAddress(msg string) Error {
	return newError(CodeUnrecognizedAddress, msg)
}

func ErrInvalidSequence(msg string) Error {
	return newError(CodeInvalidSequence, msg)
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

func NewError(code uint32, msg string) Error {
	return newError(code, msg)
}

type traceItem struct {
	msg      string
	filename string
	lineno   int
}

func (ti traceItem) String() string {
	return fmt.Sprintf("%v:%v %v", ti.filename, ti.lineno, ti.msg)
}

type sdkError struct {
	code  uint32
	msg   string
	cause error
	trace []traceItem
}

func newError(code uint32, msg string) *sdkError {
	// TODO capture stacktrace if ENV is set.
	if msg == "" {
		msg = CodeToDefaultMsg(code)
	}
	return &sdkError{
		code:  code,
		msg:   msg,
		cause: nil,
		trace: nil,
	}
}

// Implements ABCIError.
func (err *sdkError) Error() string {
	return fmt.Sprintf("Error{%d:%s,%v,%v}", err.code, err.msg, err.cause, len(err.trace))
}

// Implements ABCIError.
func (err *sdkError) ABCICode() uint32 {
	return err.code
}

// Implements ABCIError.
func (err *sdkError) ABCILog() string {
	traceLog := ""
	for _, ti := range err.trace {
		traceLog += ti.String() + "\n"
	}
	return fmt.Sprintf("msg: %v\ntrace:\n%v",
		err.msg,
		traceLog,
	)
}

// Add tracing information with msg.
func (err *sdkError) Trace(msg string) Error {
	_, fn, line, ok := runtime.Caller(1)
	if !ok {
		if fn == "" {
			fn = "<unknown>"
		}
		if line <= 0 {
			line = -1
		}
	}
	// Include file & line number & msg.
	// Do not include the whole stack trace.
	err.trace = append(err.trace, traceItem{
		filename: fn,
		lineno:   line,
		msg:      msg,
	})
	return err
}

// Add tracing information with cause and msg.
func (err *sdkError) TraceCause(cause error, msg string) Error {
	err.cause = cause
	return err.Trace(msg)
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
