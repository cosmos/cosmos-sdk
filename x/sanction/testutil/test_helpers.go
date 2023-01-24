package testutil

import (
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This file contains some functions handy for doing unit tests.

// AssertErrorContents asserts that, if contains is empty, there's no error.
// Otherwise, asserts that there is an error, and that it contains each of the provided strings.
func AssertErrorContents(t *testing.T, theError error, contains []string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(contains) == 0 {
		return assert.NoError(t, theError, msgAndArgs)
	}
	if !assert.Error(t, theError, msgAndArgs...) {
		return false
	}

	hasAll := true
	for _, expInErr := range contains {
		hasAll = assert.ErrorContains(t, theError, expInErr, msgAndArgs...) && hasAll
	}
	return hasAll
}

// didPanic safely executes the provided function and returns info about any panic it might have encountered.
func didPanic(f assert.PanicTestFunc) (didPanic bool, message interface{}, stack string) {
	didPanic = true

	defer func() {
		message = recover()
		if didPanic {
			stack = string(debug.Stack())
		}
	}()

	f()
	didPanic = false

	return
}

// AssertPanicsWithMessage asserts that the code inside the specified PanicTestFunc panics, and that
// the recovered panic message equals the expected panic message.
//
//   AssertPanicsWithMessage(t, "crazy error", func(){ GoCrazy() })
//
// PanicsWithValue requires a specific interface{} value to be provided, which can be problematic.
// PanicsWithError requires that the panic value is an error.
// This one uses fmt.Sprintf("%v", panicValue) to convert the panic recovery value to a string to test against.
func AssertPanicsWithMessage(t *testing.T, expected string, f assert.PanicTestFunc, msgAndArgs ...interface{}) bool {
	t.Helper()

	funcDidPanic, panicValue, panickedStack := didPanic(f)
	if !funcDidPanic {
		msg := fmt.Sprintf("func %#v should panic, but did not.", f)
		msg += fmt.Sprintf("\n\tExpected message:\t%q", expected)
		return assert.Fail(t, msg, msgAndArgs...)
	}
	panicMsg := fmt.Sprintf("%v", panicValue)
	if panicMsg == expected {
		return true
	}

	msg := fmt.Sprintf("func %#v panic message incorrect.", f)
	msg += fmt.Sprintf("\n\tExpected:\t%q", expected)
	msg += fmt.Sprintf("\n\t  Actual:\t%q", panicMsg)
	msg += fmt.Sprintf("\n\tPanic value:\t%#v", panicValue)
	msg += fmt.Sprintf("\n\tPanic stack:\t%s", panickedStack)
	return assert.Fail(t, msg, msgAndArgs...)
}

// RequirePanicsWithMessage asserts that the code inside the specified PanicTestFunc panics, and that
// the recovered panic message equals the expected panic message.
//
//   RequirePanicsWithMessage(t, "crazy error", func(){ GoCrazy() })
//
// PanicsWithValue requires a specific interface{} value to be provided, which can be problematic.
// PanicsWithError requires that the panic value is an error.
// This one uses fmt.Sprintf("%v", panicValue) to convert the panic recovery value to a string to test against.
//
// If the assertion fails, the test is halted.
func RequirePanicsWithMessage(t *testing.T, expected string, f assert.PanicTestFunc, msgAndArgs ...interface{}) {
	t.Helper()
	if AssertPanicsWithMessage(t, expected, f, msgAndArgs...) {
		return
	}
	t.FailNow()
}

// AssertPanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
func AssertPanicContents(t *testing.T, contains []string, f assert.PanicTestFunc, msgAndArgs ...interface{}) bool {
	t.Helper()

	funcDidPanic, panicValue, panickedStack := didPanic(f)
	panicMsg := fmt.Sprintf("%v", panicValue)

	if len(contains) == 0 {
		if !funcDidPanic {
			return true
		}
		msg := fmt.Sprintf("func %#v should not panic, but did.", f)
		msg += fmt.Sprintf("\n\tPanic message:\t%q", panicMsg)
		msg += fmt.Sprintf("\n\t  Panic value:\t%#v", panicValue)
		msg += fmt.Sprintf("\n\t  Panic stack:\t%s", panickedStack)
		return assert.Fail(t, msg, msgAndArgs...)
	}

	if !funcDidPanic {
		msg := fmt.Sprintf("func %#v should panic, but did not.", f)
		for _, exp := range contains {
			msg += fmt.Sprintf("\n\tExpected to contain:\t%q", exp)
		}
		return assert.Fail(t, msg, msgAndArgs...)
	}

	var missing []string
	for _, exp := range contains {
		if !strings.Contains(panicMsg, exp) {
			missing = append(missing, exp)
		}
	}

	if len(missing) == 0 {
		return true
	}

	msg := fmt.Sprintf("func %#v panic message incorrect.", f)
	msg += fmt.Sprintf("\n\t   Panic message:\t%q", panicMsg)
	for _, exp := range missing {
		msg += fmt.Sprintf("\n\tDoes not contain:\t%q", exp)
	}
	msg += fmt.Sprintf("\n\tPanic value:\t%#v", panicValue)
	msg += fmt.Sprintf("\n\tPanic stack:\t%s", panickedStack)
	return assert.Fail(t, msg, msgAndArgs)
}

// RequirePanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
//
// If the assertion fails, the test is halted.
func RequirePanicContents(t *testing.T, contains []string, f assert.PanicTestFunc, msgAndArgs ...interface{}) {
	t.Helper()
	if AssertPanicContents(t, contains, f, msgAndArgs...) {
		return
	}
	t.FailNow()
}

// AssertNotPanicsNoError asserts that the code inside the provided function does not panic
// and that it does not return an error.
// Returns true if it neither panics nor errors.
func AssertNotPanicsNoError(t *testing.T, f func() error, msgAndArgs ...interface{}) bool {
	t.Helper()
	var err error
	if !assert.NotPanics(t, func() { err = f() }, msgAndArgs...) {
		return false
	}
	return assert.NoError(t, err, msgAndArgs...)
}

// RequireNotPanicsNoError asserts that the code inside the provided function does not panic
// and that it does not return an error.
//
// If the assertion fails, the test is halted.
func RequireNotPanicsNoError(t *testing.T, f func() error, msgAndArgs ...interface{}) {
	t.Helper()
	if AssertNotPanicsNoError(t, f, msgAndArgs...) {
		return
	}
	t.FailNow()
}
