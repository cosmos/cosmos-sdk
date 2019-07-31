package errors

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// RootCodespace is the codespace for all errors defined in this package
const RootCodespace = "sdk"

// UndefinedCodespace when we explicitly declare no codespace
const UndefinedCodespace = "undefined"

var (
	// errInternal should never be exposed, but we reserve this code for non-specified errors
	//nolint
	errInternal = Register(UndefinedCodespace, 1, "internal")

	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode = Register(RootCodespace, 2, "tx parse error")

	// ErrInvalidSequence is used the sequence number (nonce) is incorrect
	// for the signature
	ErrInvalidSequence = Register(RootCodespace, 3, "invalid sequence")

	// ErrUnauthorized is used whenever a request without sufficient
	// authorization is handled.
	ErrUnauthorized = Register(RootCodespace, 4, "unauthorized")

	// ErrInsufficientFunds is used when the account cannot pay requested amount.
	ErrInsufficientFunds = Register(RootCodespace, 5, "insufficient funds")

	// ErrUnknownRequest to doc
	ErrUnknownRequest = Register(RootCodespace, 6, "unknown request")

	// ErrInvalidAddress to doc
	ErrInvalidAddress = Register(RootCodespace, 7, "invalid address")

	// ErrInvalidPubKey to doc
	ErrInvalidPubKey = Register(RootCodespace, 8, "invalid pubkey")

	// ErrUnknownAddress to doc
	ErrUnknownAddress = Register(RootCodespace, 9, "unknown address")

	// ErrInsufficientCoins to doc (what is the difference between ErrInsufficientFunds???)
	ErrInsufficientCoins = Register(RootCodespace, 10, "insufficient coins")

	// ErrInvalidCoins to doc
	ErrInvalidCoins = Register(RootCodespace, 11, "invalid coins")

	// ErrOutOfGas to doc
	ErrOutOfGas = Register(RootCodespace, 12, "out of gas")

	// ErrMemoTooLarge to doc
	ErrMemoTooLarge = Register(RootCodespace, 13, "memo too large")

	// ErrInsufficientFee to doc
	ErrInsufficientFee = Register(RootCodespace, 14, "insufficient fee")

	// ErrTooManySignatures to doc
	ErrTooManySignatures = Register(RootCodespace, 15, "maximum numer of signatures exceeded")

	// ErrNoSignatures to doc
	ErrNoSignatures = Register(RootCodespace, 16, "no signatures supplied")

	// ErrPanic is only set when we recover from a panic, so we know to
	// redact potentially sensitive system info
	ErrPanic = Register(UndefinedCodespace, 111222, "panic")
)

// Register returns an error instance that should be used as the base for
// creating error instances during runtime.
//
// Popular root errors are declared in this package, but extensions may want to
// declare custom codes. This function ensures that no error code is used
// twice. Attempt to reuse an error code results in panic.
//
// Use this function only during a program startup phase.
func Register(codespace string, code uint32, description string) *Error {
	// TODO - uniqueness is (codespace, code) combo
	if e := getUsed(codespace, code); e != nil {
		panic(fmt.Sprintf("error with code %d is already registered: %q", code, e.desc))
	}
	err := &Error{
		code:      code,
		codespace: codespace,
		desc:      description,
	}
	setUsed(err)
	return err
}

// usedCodes is keeping track of used codes to ensure their uniqueness. No two
// error instances should share the same (codespace, code) tuple.
var usedCodes = map[string]*Error{}

func errorID(codespace string, code uint32) string {
	return fmt.Sprintf("%s:%d", codespace, code)
}

func getUsed(codespace string, code uint32) *Error {
	return usedCodes[errorID(codespace, code)]
}

func setUsed(err *Error) {
	usedCodes[errorID(err.codespace, err.code)] = err
}

// ABCIError will resolve an error code/log from an abci result into
// an error message. If the code is registered, it will map it back to
// the canonical error, so we can do eg. ErrNotFound.Is(err) on something
// we get back from an external API.
//
// This should *only* be used in clients, not in the server side.
// The server (abci app / blockchain) should only refer to registered errors
func ABCIError(codespace string, code uint32, log string) error {
	if e := getUsed(codespace, code); e != nil {
		return Wrap(e, log)
	}
	// This is a unique error, will never match on .Is()
	// Use Wrap here to get a stack trace
	return Wrap(&Error{codespace: codespace, code: code, desc: "unknown"}, log)
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
	codespace string
	code      uint32
	desc      string
}

func (e Error) Error() string {
	return e.desc
}

func (e Error) ABCICode() uint32 {
	return e.code
}

func (e Error) Codespace() string {
	return e.codespace
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
