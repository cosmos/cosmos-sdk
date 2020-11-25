package rosetta

// errors.go contains all the errors returned by the adapter implementation
// plus some extra utilities to parse those errors

import (
	"fmt"

	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// Error defines an error that can be converted to a Rosetta API error.
type Error interface {
	error
	// RosettaError returns the error as a Rosetta API error.
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
	status, ok := grpcstatus.FromError(err)
	if !ok {
		return WrapError(ErrUnknown, err.Error())
	}
	switch status.Code() {
	case grpccodes.NotFound:
		return WrapError(ErrNotFound, status.Message())
	case grpccodes.FailedPrecondition:
		return WrapError(ErrBadArgument, status.Message())
	case grpccodes.InvalidArgument:
		return WrapError(ErrBadArgument, status.Message())
	case grpccodes.Internal:
		return WrapError(ErrInternal, status.Message())
	default:
		return WrapError(ErrUnknown, status.Message())
	}
}

// Default error list
var (
	// ErrUnknown defines an unknown error, if this is returned it means
	// the library is ignoring an error
	ErrUnknown = NewError(0, "unknown", false)
	// ErrBadArgument is returned when the request is malformed
	ErrBadArgument = NewError(400, "bad argument", false)
	// ErrNotFound is returned when the required object was not found
	// retry is set to true because something that is not found now
	// might be found later, example: a TX
	ErrNotFound = NewError(404, "not found", true)
	// ErrInternal is returned when the node is experiencing internal errors
	ErrInternal = NewError(500, "internal error", false)
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
	ErrInvalidPubkey      = NewError(8, "invalid pubkey", false)
	ErrInterpreting       = NewError(1, "error interpreting data from node", false)
	ErrInvalidAddress     = NewError(7, "invalid address", false)
	ErrInvalidMemo        = NewError(11, "invalid memo", false)
	ErrInvalidOperation   = NewError(4, "invalid operation", false)
	ErrInvalidRequest     = NewError(6, "invalid request", false)
	ErrInvalidTransaction = NewError(5, "invalid transaction", false)
)

// AllowedErrors lists all the rosetta errors
// returned by the API, it's required for the
// 'allow' network information endpoint
var AllowedErrors = Errors{
	ErrUnknown,
	ErrNotFound,
	ErrBadArgument,
	ErrInternal,
	ErrBadGateway,
	ErrOffline,
	ErrNetworkNotSupported,
	ErrCodec,
	ErrNotImplemented,
	ErrUnsupportedCurve,
	ErrInvalidPubkey,
	ErrInvalidAddress,
	ErrInvalidMemo,
	ErrInvalidOperation,
	ErrInvalidRequest,
	ErrInvalidTransaction,
}
