package types

import (
	"fmt"
	"github.com/tendermint/go-crypto"
	"runtime"
)

type CodeType uint32

func (code CodeType) IsOK() bool {
	if code == CodeOK {
		return true
	} else {
		return false
	}
}

const (
	// ABCI Response Codes
	// Base SDK reserves 0 ~ 99.
	CodeOK                  CodeType = 0
	CodeInternal            CodeType = 1
	CodeTxParse             CodeType = 2
	CodeBadNonce            CodeType = 3
	CodeUnauthorized        CodeType = 4
	CodeInsufficientFunds   CodeType = 5
	CodeUnknownRequest      CodeType = 6
	CodeUnrecognizedAddress CodeType = 7
	CodeInvalidSequence     CodeType = 8
)

// NOTE: Don't stringer this, we'll put better messages in later.
func CodeToDefaultMsg(code CodeType) string {
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

func ErrUnrecognizedAddress(addr crypto.Address) Error {
	return newError(CodeUnrecognizedAddress, addr.String())
}

func ErrInvalidSequence(msg string) Error {
	return newError(CodeInvalidSequence, msg)
}

//----------------------------------------
// Error & sdkError

type Error interface {
	Error() string
	ABCICode() CodeType
	ABCILog() string
	Trace(msg string) Error
	TraceCause(cause error, msg string) Error
	Cause() error
	Result() Result
}

func NewError(code CodeType, msg string) Error {
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
	code   CodeType
	msg    string
	cause  error
	traces []traceItem
}

func newError(code CodeType, msg string) *sdkError {
	// TODO capture stacktrace if ENV is set.
	if msg == "" {
		msg = CodeToDefaultMsg(code)
	}
	return &sdkError{
		code:   code,
		msg:    msg,
		cause:  nil,
		traces: nil,
	}
}

// Implements ABCIError.
func (err *sdkError) Error() string {
	return fmt.Sprintf("Error{%d:%s,%v,%v}", err.code, err.msg, err.cause, len(err.traces))
}

// Implements ABCIError.
func (err *sdkError) ABCICode() CodeType {
	return err.code
}

// Implements ABCIError.
func (err *sdkError) ABCILog() string {
	traceLog := ""
	for _, ti := range err.traces {
		traceLog += ti.String() + "\n"
	}
	return fmt.Sprintf("msg: %v\ntrace:\n%v",
		err.msg,
		traceLog,
	)
}

// Add tracing information with msg.
func (err *sdkError) Trace(msg string) Error {
	return err.doTrace(msg, 2)
}

// Add tracing information with cause and msg.
func (err *sdkError) TraceCause(cause error, msg string) Error {
	err.cause = cause
	return err.doTrace(msg, 2)
}

func (err *sdkError) doTrace(msg string, n int) Error {
	_, fn, line, ok := runtime.Caller(n)
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
	err.traces = append(err.traces, traceItem{
		filename: fn,
		lineno:   line,
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
