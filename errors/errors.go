package errors

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

var (
	// errInternal should never be exposed, but we reserve this code for non-specified errors
	//nolint
	errInternal = Register(1, "internal")

	// ErrUnauthorized is used whenever a request without sufficient
	// authorization is handled.
	ErrUnauthorized = Register(2, "unauthorized")

	// ErrNotFound is used when a requested operation cannot be completed
	// due to missing data.
	ErrNotFound = Register(3, "not found")

	// ErrMsg is returned whenever an event is invalid and cannot be
	// handled.
	ErrMsg = Register(4, "invalid message")

	// ErrModel is returned whenever a message is invalid and cannot
	// be used (ie. persisted).
	ErrModel = Register(5, "invalid model")

	// ErrDuplicate is returned when there is a record already that has the same
	// unique key/index used
	ErrDuplicate = Register(6, "duplicate")

	// ErrHuman is returned when application reaches a code path which should not
	// ever be reached if the code was written as expected by the framework
	ErrHuman = Register(7, "coding error")

	// ErrImmutable is returned when something that is considered immutable
	// gets modified
	ErrImmutable = Register(8, "cannot be modified")

	// ErrEmpty is returned when a value fails a not empty assertion
	ErrEmpty = Register(9, "value is empty")

	// ErrState is returned when an object is in invalid state
	ErrState = Register(10, "invalid state")

	// ErrType is returned whenever the type is not what was expected
	ErrType = Register(11, "invalid type")

	// ErrAmount is returned when processed amount is invalid.
	ErrAmount = Register(13, "invalid amount")

	// ErrInput stands for general input problems indication
	ErrInput = Register(14, "invalid input")

	// ErrExpired stands for expired entities, normally has to do with block height expirations
	ErrExpired = Register(15, "expired")

	// ErrOverflow s returned when a computation cannot be completed
	// because the result value exceeds the type.
	ErrOverflow = Register(16, "an operation cannot be completed due to value overflow")

	// ErrCurrency is returned whenever an operation cannot be completed
	// due to a currency issues.
	ErrCurrency = Register(17, "currency")

	// ErrMetadata is returned whenever a weave.Metadata payload is invalid.
	ErrMetadata = Register(18, "metadata")

	// ErrSchema is returned whenever an operation cannot be completed due
	// to an object schema version issue.
	ErrSchema = Register(19, "schema")

	// ErrDatabase is returned whenever the underlying kvstore fails to
	// process raw bytes (get/set/delete/write)
	ErrDatabase = Register(20, "database")

	// ErrDeleted is returned whenever a deleted object version is accessed.
	ErrDeleted = Register(21, "content deleted")

	// ErrIteratorDone is returned when an iterator hits the end of the data source.
	ErrIteratorDone = Register(22, "iterator done")

	// ErrNetwork is returned on network failure (only for client libraries)
	ErrNetwork = Register(100200, "network")

	// ErrTimeout is returned on context timeout (only for client libraries)
	ErrTimeout = Register(100300, "timeout")

	// ErrPanic is only set when we recover from a panic, so we know to
	// redact potentially sensitive system info
	ErrPanic = Register(111222, "panic")
)

// Register returns an error instance that should be used as the base for
// creating error instances during runtime.
//
// Popular root errors are declared in this package, but extensions may want to
// declare custom codes. This function ensures that no error code is used
// twice. Attempt to reuse an error code results in panic.
//
// Use this function only during a program startup phase.
func Register(code uint32, description string) *Error {
	if e, ok := usedCodes[code]; ok {
		panic(fmt.Sprintf("error with code %d is already registered: %q", code, e.desc))
	}
	err := &Error{
		code: code,
		desc: description,
	}
	usedCodes[err.code] = err
	return err
}

// usedCodes is keeping track of used codes to ensure their uniqueness. No two
// error instances should share the same error code.
var usedCodes = map[uint32]*Error{
	// Register multi error code so that it cannot be used.
	multiErrorABCICode: nil,
}

// ABCIError will resolve an error code/log from an abci result into
// an error message. If the code is registered, it will map it back to
// the canonical error, so we can do eg. ErrNotFound.Is(err) on something
// we get back from an external API.
//
// This should *only* be used in clients, not in the server side.
// The server (abci app / blockchain) should only refer to registered errors
func ABCIError(code uint32, log string) error {
	if e, ok := usedCodes[code]; ok {
		return Wrap(e, log)
	}
	// This is a unique error, will never match on .Is()
	// Use Wrap here to get a stack trace
	return Wrap(&Error{code: code}, log)
}

// Error represents a root error.
//
// Weave framework is using root error to categorize issues. Each instance
// created during the runtime should wrap one of the declared root errors. This
// allows error tests and returning all errors to the client in a safe manner.
//
// All popular root errors are declared in this package. If an extension has to
// declare a custom root error, always use Register function to ensure
// error code uniqueness.
type Error struct {
	code uint32
	desc string
}

func (e Error) Error() string {
	return e.desc
}

func (e Error) ABCICode() uint32 {
	return e.code
}

// Is check if given error instance is of a given kind/type. This involves
// unwrapping given error using the Cause method if available.
func (kind *Error) Is(err error) bool {
	// Reflect usage is necessary to correctly compare with
	// a nil implementation of an error.
	if kind == nil {
		return isNilErr(err)
	}

	for {
		if err == kind {
			return true
		}

		// If this is a collection of errors, this function must return
		// true if at least one from the group match.
		if u, ok := err.(unpacker); ok {
			for _, e := range u.Unpack() {
				if kind.Is(e) {
					return true
				}
			}
		}

		if c, ok := err.(causer); ok {
			err = c.Cause()
		} else {
			return false
		}
	}
}

func isNilErr(err error) bool {
	// Reflect usage is necessary to correctly compare with
	// a nil implementation of an error.
	if err == nil {
		return true
	}
	if reflect.ValueOf(err).Kind() == reflect.Struct {
		return false
	}
	return reflect.ValueOf(err).IsNil()
}

// Wrap extends given error with an additional information.
//
// If the wrapped error does not provide ABCICode method (ie. stdlib errors),
// it will be labeled as internal error.
//
// If err is nil, this returns nil, avoiding the need for an if statement when
// wrapping a error returned at the end of a function
func Wrap(err error, description string) error {
	if err == nil {
		return nil
	}

	// If this error does not carry the stacktrace information yet, attach
	// one. This should be done only once per error at the lowest frame
	// possible (most inner wrap).
	if stackTrace(err) == nil {
		err = errors.WithStack(err)
	}

	return &wrappedError{
		parent: err,
		msg:    description,
	}
}

// Wrapf extends given error with an additional information.
//
// This function works like Wrap function with additional functionality of
// formatting the input as specified.
func Wrapf(err error, format string, args ...interface{}) error {
	desc := fmt.Sprintf(format, args...)
	return Wrap(err, desc)
}

type wrappedError struct {
	// This error layer description.
	msg string
	// The underlying error that triggered this one.
	parent error
}

func (e *wrappedError) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.parent.Error())
}

func (e *wrappedError) Cause() error {
	return e.parent
}

// Recover captures a panic and stop its propagation. If panic happens it is
// transformed into a ErrPanic instance and assigned to given error. Call this
// function using defer in order to work as expected.
func Recover(err *error) {
	if r := recover(); r != nil {
		*err = Wrapf(ErrPanic, "%v", r)
	}
}

// WithType is a helper to augment an error with a corresponding type message
func WithType(err error, obj interface{}) error {
	return Wrap(err, fmt.Sprintf("%T", obj))
}

// causer is an interface implemented by an error that supports wrapping. Use
// it to test if an error wraps another error instance.
type causer interface {
	Cause() error
}

type unpacker interface {
	Unpack() []error
}
