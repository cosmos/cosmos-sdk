package errors

import (
	"fmt"
)

// UndefinedCodespace when we explicitly declare no codespace
const UndefinedCodespace = "undefined"

var (

	// ErrStopIterating is used to break out of an iteration
	ErrStopIterating = Register(UndefinedCodespace, 2, "stop iterating")

	// ErrPanic should only be set when we recovering from a panic
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
	if e := getUsed(codespace, code); e != nil {
		panic(fmt.Sprintf("error with code %d is already registered: %q", code, e.desc))
	}

	err := &Error{codespace: codespace, code: code, desc: description}
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
		return fmt.Errorf("%s: %w", log, e)
	}
	// This is a unique error, will never match on .Is()
	// Use Wrap here to get a stack trace
	return fmt.Errorf("%s: %w", log, &Error{codespace: codespace, code: code, desc: "unknown"})
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

// New is an alias for Register.
func New(codespace string, code uint32, desc string) *Error {
	return Register(codespace, code, desc)
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

	return fmt.Errorf("%s: %w", description, err)
}

// Wrapf extends given error with an additional information.
//
// This function works like Wrap function with additional functionality of
// formatting the input as specified.
func Wrapf(err error, format string, args ...interface{}) error {
	desc := fmt.Sprintf(format, args...)
	return Wrap(err, desc)
}
