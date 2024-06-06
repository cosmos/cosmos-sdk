package errors

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	// SuccessABCICode declares an ABCI response use 0 to signal that the
	// processing was successful and no error is returned.
	SuccessABCICode uint32 = 0

	// All unclassified errors that do not provide an ABCI code are clubbed
	// under an internal error code and a generic message instead of
	// detailed error string.
	internalABCICodespace        = UndefinedCodespace
	internalABCICode      uint32 = 1
)

// ABCIInfo returns the ABCI error information as consumed by the tendermint
// client. Returned codespace, code, and log message should be used as a ABCI response.
// Any error that does not provide ABCICode information is categorized as error
// with code 1, codespace UndefinedCodespace
// When not running in a debug mode all messages of errors that do not provide
// ABCICode information are replaced with generic "internal error". Errors
// without an ABCICode information as considered internal.
func ABCIInfo(err error, debug bool) (codespace string, code uint32, log string) {
	if errIsNil(err) {
		return "", SuccessABCICode, ""
	}

	encode := defaultErrEncoder
	if debug {
		encode = debugErrEncoder
	}

	code, space := abciInfo(err)
	return space, code, encode(err)
}

// The debugErrEncoder encodes the error with a stacktrace.
func debugErrEncoder(err error) string {
	return fmt.Sprintf("%+v", err)
}

func defaultErrEncoder(err error) string {
	return err.Error()
}

// abciInfo tests if given error contains an ABCI code and returns the value of
// it if available. This function is testing for the causer interface as well
// and unwraps the error.
func abciInfo(err error) (code uint32, codespace string) {
	if errIsNil(err) {
		return SuccessABCICode, ""
	}

	var customErr *Error

	if errors.As(err, &customErr) {
		code = customErr.ABCICode()
		codespace = customErr.Codespace()
	} else {
		code = internalABCICode
		codespace = internalABCICodespace
	}

	return
}

// errIsNil returns true if value represented by the given error is nil.
//
// Most of the time a simple == check is enough. There is a very narrowed
// spectrum of cases (mostly in tests) where a more sophisticated check is
// required.
func errIsNil(err error) bool {
	if err == nil {
		return true
	}
	if val := reflect.ValueOf(err); val.Kind() == reflect.Ptr {
		return val.IsNil()
	}
	return false
}
