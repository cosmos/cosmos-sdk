// errors.go contains all the errors returned by the adapter implementation
// plus some extra utilities to parse those errors
package rosetta

import (
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// Error defines an error that can be converted to a rosetta error
type Error interface {
	error
	// RosettaError returns the error as a rosetta error
	RosettaError() *types.Error
}

// Errors simply defines an array of Error
type Errors []Error

// RosettaErrors converts all the errors into rosetta ones
func (e Errors) RosettaErrors() []*types.Error {
	rosErrs := make([]*types.Error, len(e))
	for i, err := range e {
		rosErrs[i] = err.RosettaError()
	}
	return rosErrs
}

type errorWrapper struct {
	err *types.Error
}

func (e errorWrapper) Error() string {
	return fmt.Sprintf("rosetta: %d %s", e.err.Code, e.err.Message)
}

func (e errorWrapper) RosettaError() *types.Error {
	return e.err
}

// Is implements the errors.Is interface to verify
// if the given error matches one in the library
func (e errorWrapper) Is(target error) bool {
	err, ok := target.(Error)
	if !ok {
		return false
	}
	return err.RosettaError().Code == e.err.Code
}

// NewError instantiates a new rosetta error
func NewError(code int32, message string, retry bool) Error {
	return errorWrapper{
		err: &types.Error{
			Code:        code,
			Message:     message,
			Description: nil,
			Retriable:   retry,
			Details:     nil,
		},
	}
}

// WrapError wraps the rosetta error with additional context
func WrapError(err Error, msg string) Error {
	e := err.RosettaError()
	wrappedMsg := fmt.Sprintf("%s: %s", e.Message, msg)
	return NewError(e.Code, wrappedMsg, e.Retriable)
}

// ToRosettaError attempts to converting an error into a rosetta
// error, if the error cannot be converted it will be parsed as unknown
func ToRosettaError(err error) *types.Error {
	rosErr, ok := err.(Error)
	if !ok {
		return WrapError(ErrUnknown, err.Error()).RosettaError()
	}
	return rosErr.RosettaError()
}

// FromGRPCToRosettaError converts a gRPC error to rosetta error
func FromGRPCToRosettaError(err error) Error {
	return errorWrapper{
		err: &types.Error{
			Code:        0,
			Message:     err.Error(),
			Description: nil,
			Retriable:   false,
			Details:     nil,
		},
	}
}

// Default error list
var (
	// ErrUnknown defines an unknown error, if this is returned it means
	// the library is ignoring an error
	ErrUnknown = NewError(0, "unknown", true)
	// ErrBadArgument is returned when the request is malformed
	ErrBadArgument = NewError(400, "bad argument", false)
	// ErrBadGateway is returned when there are problems interacting with the nodes
	ErrBadGateway = NewError(502, "bad gateway", true)
	// ErrOffline is returned when there is an attempt to query an endpoint in offline mode
	ErrOffline = NewError(1, "cannot query endpoint in offline mode", false)
	// ErrNetworkNotSupported is returned when there is an attempt to query a network which is not supported
	ErrNetworkNotSupported = NewError(2, "network is not supported", false)
	// ErrCodec is returned when there's an error while marshalling or unmarshalling data
	ErrCodec = NewError(3, "encode/decode error", true)
	// ErrNotImplemented is returned when a method is not implemented yet
	ErrNotImplemented = NewError(14, "not implemented", false)
	// ErrUnsupportedCurve is returned when the curve specified is not supported
	ErrUnsupportedCurve = NewError(15, "unsupported curve, expected secp256k1", false)
	// ErrInvalidPubkey is returned when the public key is invalid
	ErrInvalidPubkey = NewError(8, "invalid pubkey", false)
)

// AllowedErrors lists all the rosetta allowed errors
var AllowedErrors = Errors{
	ErrUnknown,
	ErrBadArgument,
	ErrBadGateway,
	ErrOffline,
	ErrNetworkNotSupported,
	ErrCodec,
	ErrNotImplemented,
	ErrUnsupportedCurve,
	ErrInvalidPubkey,
}
