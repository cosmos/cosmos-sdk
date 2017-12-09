//nolint
package errors

import (
	"fmt"
	"reflect"
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
	return WithMessage(msg, errUnknownTxType, CodeTypeUnknownRequest)
}
func IsUnknownTxTypeErr(err error) bool {
	return IsSameError(errUnknownTxType, err)
}

func ErrInvalidFormat(expected string, tx interface{}) TMError {
	msg := fmt.Sprintf("%T not %s", unwrap(tx), expected)
	return WithMessage(msg, errInvalidFormat, CodeTypeUnknownRequest)
}
func IsInvalidFormatErr(err error) bool {
	return IsSameError(errInvalidFormat, err)
}

func ErrUnknownModule(mod string) TMError {
	return WithMessage(mod, errUnknownModule, CodeTypeUnknownRequest)
}
func IsUnknownModuleErr(err error) bool {
	return IsSameError(errUnknownModule, err)
}

func ErrUnknownKey(mod string) TMError {
	return WithMessage(mod, errUnknownKey, CodeTypeUnknownRequest)
}
func IsUnknownKeyErr(err error) bool {
	return IsSameError(errUnknownKey, err)
}

func ErrInternal(msg string) TMError {
	return New(msg, CodeTypeInternalErr)
}

// IsInternalErr matches any error that is not classified
func IsInternalErr(err error) bool {
	return HasErrorCode(err, CodeTypeInternalErr)
}

func ErrDecoding() TMError {
	return WithCode(errDecoding, CodeTypeEncodingErr)
}
func IsDecodingErr(err error) bool {
	return IsSameError(errDecoding, err)
}

func ErrUnauthorized() TMError {
	return WithCode(errUnauthorized, CodeTypeUnauthorized)
}

// IsUnauthorizedErr is generic helper for any unauthorized errors,
// also specific sub-types
func IsUnauthorizedErr(err error) bool {
	return HasErrorCode(err, CodeTypeUnauthorized)
}

func ErrMissingSignature() TMError {
	return WithCode(errMissingSignature, CodeTypeUnauthorized)
}
func IsMissingSignatureErr(err error) bool {
	return IsSameError(errMissingSignature, err)
}

func ErrTooLarge() TMError {
	return WithCode(errTooLarge, CodeTypeEncodingErr)
}
func IsTooLargeErr(err error) bool {
	return IsSameError(errTooLarge, err)
}
