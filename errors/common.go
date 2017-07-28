//nolint
package errors

import (
	"fmt"
	"reflect"

	abci "github.com/tendermint/abci/types"
)

var (
	errDecoding         = fmt.Errorf("Error decoding input")
	errUnauthorized     = fmt.Errorf("Unauthorized")
	errTooLarge         = fmt.Errorf("Input size too large")
	errMissingSignature = fmt.Errorf("Signature missing")
	errUnknownTxType    = fmt.Errorf("Tx type unknown")
	errInvalidFormat    = fmt.Errorf("Invalid format")
	errUnknownModule    = fmt.Errorf("Unknown module")
	errUnknownKey       = fmt.Errorf("Unknown key")

	internalErr    = abci.CodeType_InternalError
	encodingErr    = abci.CodeType_EncodingError
	unauthorized   = abci.CodeType_Unauthorized
	unknownRequest = abci.CodeType_UnknownRequest
	unknownAddress = abci.CodeType_BaseUnknownAddress
)

// some crazy reflection to unwrap any generated struct.
func unwrap(i interface{}) interface{} {
	v := reflect.ValueOf(i)
	m := v.MethodByName("Unwrap")
	if m.IsValid() {
		out := m.Call(nil)
		if len(out) == 1 {
			return out[0].Interface()
		}
	}
	return i
}

func ErrUnknownTxType(tx interface{}) TMError {
	msg := fmt.Sprintf("%T", unwrap(tx))
	return WithMessage(msg, errUnknownTxType, unknownRequest)
}
func IsUnknownTxTypeErr(err error) bool {
	return IsSameError(errUnknownTxType, err)
}

func ErrInvalidFormat(expected string, tx interface{}) TMError {
	msg := fmt.Sprintf("%T not %s", unwrap(tx), expected)
	return WithMessage(msg, errInvalidFormat, unknownRequest)
}
func IsInvalidFormatErr(err error) bool {
	return IsSameError(errInvalidFormat, err)
}

func ErrUnknownModule(mod string) TMError {
	return WithMessage(mod, errUnknownModule, unknownRequest)
}
func IsUnknownModuleErr(err error) bool {
	return IsSameError(errUnknownModule, err)
}

func ErrUnknownKey(mod string) TMError {
	return WithMessage(mod, errUnknownKey, unknownRequest)
}
func IsUnknownKeyErr(err error) bool {
	return IsSameError(errUnknownKey, err)
}

func ErrInternal(msg string) TMError {
	return New(msg, internalErr)
}

// IsInternalErr matches any error that is not classified
func IsInternalErr(err error) bool {
	return HasErrorCode(err, internalErr)
}

func ErrDecoding() TMError {
	return WithCode(errDecoding, encodingErr)
}
func IsDecodingErr(err error) bool {
	return IsSameError(errDecoding, err)
}

func ErrUnauthorized() TMError {
	return WithCode(errUnauthorized, unauthorized)
}

// IsUnauthorizedErr is generic helper for any unauthorized errors,
// also specific sub-types
func IsUnauthorizedErr(err error) bool {
	return HasErrorCode(err, unauthorized)
}

func ErrMissingSignature() TMError {
	return WithCode(errMissingSignature, unauthorized)
}
func IsMissingSignatureErr(err error) bool {
	return IsSameError(errMissingSignature, err)
}

func ErrTooLarge() TMError {
	return WithCode(errTooLarge, encodingErr)
}
func IsTooLargeErr(err error) bool {
	return IsSameError(errTooLarge, err)
}
