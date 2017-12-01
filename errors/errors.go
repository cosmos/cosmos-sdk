package errors

import "fmt"

const (
	// ABCI Response Codes
	CodeInternalError     = 1
	CodeTxParseError      = 2
	CodeBadNonce          = 3
	CodeUnauthorized      = 4
	CodeInsufficientFunds = 5
	CodeUnknownRequest    = 6
)

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultLog(code uint32) string {
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
	default:
		return fmt.Sprintf("Unknown code %d", code)
	}
}

//--------------------------------------------------------------------------------
// All errors are created via constructors so as to enable us to hijack them
// and inject stack traces if we really want to.

func InternalError(log string) sdkError {
	return newSDKError(CodeInternalError, log)
}

func TxParseError(log string) sdkError {
	return newSDKError(CodeTxParseError, log)
}

func BadNonce(log string) sdkError {
	return newSDKError(CodeBadNonce, log)
}

func Unauthorized(log string) sdkError {
	return newSDKError(CodeUnauthorized, log)
}

func InsufficientFunds(log string) sdkError {
	return newSDKError(CodeInsufficientFunds, log)
}

func UnknownRequest(log string) sdkError {
	return newSDKError(CodeUnknownRequest, log)
}

//----------------------------------------

type ABCIError interface {
	ABCICode() uint32
	ABCILog() string
}

/*

	This struct is intended to be used with pkg/errors.

	Usage:

	```
		import sdk "github.com/cosmos/cosmos-sdk"
		import "github.com/pkg/errors"

		var err = <some causal error>
		if err != nil {
			err = sdk.InternalError("").WithCause(err)
			err = errors.Wrap(err, "Captured the stack!")
			return err
		}
	```

	Then, to get the ABCI code/log:

	1. Check if err.(ABCIError)
	2. Check if err.(causer).Cause().(ABCIError)

*/
type sdkError struct {
	code  uint32
	log   string
	cause error
	// TODO stacktrace
}

func newSDKError(code uint32, log string) sdkError {
	// TODO capture stacktrace if ENV is set
	if log == "" {
		log = codeToDefaultLog(code)
	}
	return sdkError{
		code: code,
		log:  log,
	}
}

func (err sdkError) Error() string {
	return fmt.Sprintf("SDKError{%d: %s}", err.code, err.log)
}

// Implements ABCIError
func (err sdkError) ABCICode() uint32 {
	return err.code
}

// Implements ABCIError
func (err sdkError) ABCILog() string {
	return err.log
}

func (err sdkError) Cause() error {
	return err.cause
}

func (err sdkError) WithCause(cause error) sdkError {
	copy := err
	copy.cause = cause
	return copy
}
