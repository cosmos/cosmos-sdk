package errors

import (
	"fmt"
	"reflect"

	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	// SuccessABCICode declares an ABCI response use 0 to signal that the
	// processing was successful and no error is returned.
	SuccessABCICode = 0

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

	return abciCodespace(err), abciCode(err), encode(err)
}

// ResponseCheckTx returns an ABCI ResponseCheckTx object with fields filled in
// from the given error and gas values.
func ResponseCheckTx(err error, gw, gu uint64, debug bool) abci.ResponseCheckTx {
	space, code, log := ABCIInfo(err, debug)
	return abci.ResponseCheckTx{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: int64(gw),
		GasUsed:   int64(gu),
	}
}

// ResponseCheckTxWithEvents returns an ABCI ResponseCheckTx object with fields filled in
// from the given error, gas values and events.
func ResponseCheckTxWithEvents(err error, gw, gu uint64, events []abci.Event, debug bool) abci.ResponseCheckTx {
	space, code, log := ABCIInfo(err, debug)
	return abci.ResponseCheckTx{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: int64(gw),
		GasUsed:   int64(gu),
		Events:    events,
	}
}

// ResponseDeliverTx returns an ABCI ResponseDeliverTx object with fields filled in
// from the given error and gas values.
func ResponseDeliverTx(err error, gw, gu uint64, debug bool) abci.ResponseDeliverTx {
	space, code, log := ABCIInfo(err, debug)
	return abci.ResponseDeliverTx{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: int64(gw),
		GasUsed:   int64(gu),
	}
}

// ResponseDeliverTxWithEvents returns an ABCI ResponseDeliverTx object with fields filled in
// from the given error, gas values and events.
func ResponseDeliverTxWithEvents(err error, gw, gu uint64, events []abci.Event, debug bool) abci.ResponseDeliverTx {
	space, code, log := ABCIInfo(err, debug)
	return abci.ResponseDeliverTx{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: int64(gw),
		GasUsed:   int64(gu),
		Events:    events,
	}
}

// QueryResult returns a ResponseQuery from an error. It will try to parse ABCI
// info from the error.
func QueryResult(err error) abci.ResponseQuery {
	space, code, log := ABCIInfo(err, false)
	return abci.ResponseQuery{
		Codespace: space,
		Code:      code,
		Log:       log,
	}
}

// The debugErrEncoder encodes the error with a stacktrace.
func debugErrEncoder(err error) string {
	return fmt.Sprintf("%+v", err)
}

// The defaultErrEncoder applies Redact on the error before encoding it with its internal error message.
func defaultErrEncoder(err error) string {
	return Redact(err).Error()
}

type coder interface {
	ABCICode() uint32
}

// abciCode test if given error contains an ABCI code and returns the value of
// it if available. This function is testing for the causer interface as well
// and unwraps the error.
func abciCode(err error) uint32 {
	if errIsNil(err) {
		return SuccessABCICode
	}

	for {
		if c, ok := err.(coder); ok {
			return c.ABCICode()
		}

		if c, ok := err.(causer); ok {
			err = c.Cause()
		} else {
			return internalABCICode
		}
	}
}

type codespacer interface {
	Codespace() string
}

// abciCodespace tests if given error contains a codespace and returns the value of
// it if available. This function is testing for the causer interface as well
// and unwraps the error.
func abciCodespace(err error) string {
	if errIsNil(err) {
		return ""
	}

	for {
		if c, ok := err.(codespacer); ok {
			return c.Codespace()
		}

		if c, ok := err.(causer); ok {
			err = c.Cause()
		} else {
			return internalABCICodespace
		}
	}
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

var errPanicWithMsg = Wrapf(ErrPanic, "panic message redacted to hide potentially sensitive system info")

// Redact replaces an error that is not initialized as an ABCI Error with a
// generic internal error instance. If the error is an ABCI Error, that error is
// simply returned.
func Redact(err error) error {
	if ErrPanic.Is(err) {
		return errPanicWithMsg
	}
	if abciCode(err) == internalABCICode {
		return errInternal
	}

	return err
}
