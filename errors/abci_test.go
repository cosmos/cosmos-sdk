package errors

import (
	"fmt"
	"io"
	"testing"
)

func TestABCInfo(t *testing.T) {
	cases := map[string]struct {
		err       error
		debug     bool
		wantCode  uint32
		wantSpace string
		wantLog   string
	}{
		"plain SDK error": {
			err:       ErrUnauthorized,
			debug:     false,
			wantLog:   "unauthorized",
			wantCode:  ErrUnauthorized.code,
			wantSpace: testCodespace,
		},
		"wrapped SDK error": {
			err:       fmt.Errorf("bar: %w", fmt.Errorf("foo: %w", ErrUnauthorized)),
			debug:     false,
			wantLog:   "bar: foo: unauthorized",
			wantCode:  ErrUnauthorized.code,
			wantSpace: testCodespace,
		},
		"nil is empty message": {
			err:       nil,
			debug:     false,
			wantLog:   "",
			wantCode:  0,
			wantSpace: "",
		},
		"nil SDK error is not an error": {
			err:       (*Error)(nil),
			debug:     false,
			wantLog:   "",
			wantCode:  0,
			wantSpace: "",
		},
		"stdlib returns error message in debug mode": {
			err:       io.EOF,
			debug:     true,
			wantLog:   "EOF",
			wantCode:  1,
			wantSpace: UndefinedCodespace,
		},
		// "wrapped stdlib is a full message in debug mode": {
		// 	err:      fmt.Errorf("cannot read file: %w", io.EOF),
		// 	debug:    true,
		// 	wantLog:  "cannot read file: EOF",
		// 	wantCode: 1,
		// },
		"custom error": {
			err:       customErr{},
			debug:     false,
			wantLog:   "custom",
			wantCode:  999,
			wantSpace: "extern",
		},
		"custom error in debug mode": {
			err:       customErr{},
			debug:     true,
			wantLog:   "custom",
			wantCode:  999,
			wantSpace: "extern",
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			space, code, log := ABCIInfo(tc.err, tc.debug)
			if space != tc.wantSpace {
				t.Errorf("%s: expected space %s, got %s", testName, tc.wantSpace, space)
			}
			if code != tc.wantCode {
				t.Errorf("%s: expected code %d, got %d", testName, tc.wantCode, code)
			}
			if log != tc.wantLog {
				t.Errorf("%s: expected log %s, got %s", testName, tc.wantLog, log)
			}
		})
	}
}

func TestABCIInfoHidesStacktrace(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", ErrUnauthorized)
	_, _, log := ABCIInfo(err, false)
	if log != "wrapped: unauthorized" {
		t.Errorf("expected log %s, got %s", "wrapped: unauthorized", log)
	}
}

func TestABCIInfoSerializeErr(t *testing.T) {
	var (
		// Create errors with stacktrace for equal comparison.
		myErrDecode = fmt.Errorf("test, %w", ErrTxDecode)
		myErrAddr   = fmt.Errorf("tester, %w", ErrInvalidAddress)
		myPanic     = ErrPanic
	)

	specs := map[string]struct {
		src   error
		debug bool
		exp   string
	}{
		"single error": {
			src:   myErrDecode,
			debug: false,
			exp:   "test: tx parse error",
		},
		"second error": {
			src:   myErrAddr,
			debug: false,
			exp:   "tester: invalid address",
		},
		"single error with debug": {
			src:   myErrDecode,
			debug: true,
			exp:   fmt.Sprintf("%+v", myErrDecode),
		},
		"do not redact in debug encoder": {
			src:   myPanic,
			debug: true,
			exp:   fmt.Sprintf("%+v", myPanic),
		},
	}
	for msg, spec := range specs {
		spec := spec
		_, _, log := ABCIInfo(spec.src, spec.debug)
		if log != spec.exp {
			t.Errorf("%s: expected log %s, got %s", msg, spec.exp, log)
		}
	}
}

// customErr is a custom implementation of an error that provides an ABCICode
// method.
type customErr struct{}

func (customErr) Codespace() string { return "extern" }

func (customErr) ABCICode() uint32 { return 999 }

func (customErr) Error() string { return "custom" }
