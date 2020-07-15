package errors

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
			wantSpace: RootCodespace,
		},
		"wrapped SDK error": {
			err:       Wrap(Wrap(ErrUnauthorized, "foo"), "bar"),
			debug:     false,
			wantLog:   "unauthorized: foo: bar",
			wantCode:  ErrUnauthorized.code,
			wantSpace: RootCodespace,
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
		"stdlib is generic message": {
			err:       io.EOF,
			debug:     false,
			wantLog:   "internal",
			wantCode:  1,
			wantSpace: UndefinedCodespace,
		},
		"stdlib returns error message in debug mode": {
			err:       io.EOF,
			debug:     true,
			wantLog:   "EOF",
			wantCode:  1,
			wantSpace: UndefinedCodespace,
		},
		"wrapped stdlib is only a generic message": {
			err:       Wrap(io.EOF, "cannot read file"),
			debug:     false,
			wantLog:   "internal",
			wantCode:  1,
			wantSpace: UndefinedCodespace,
		},
		// This is hard to test because of attached stacktrace. This
		// case is tested in an another test.
		//"wrapped stdlib is a full message in debug mode": {
		//	err:      Wrap(io.EOF, "cannot read file"),
		//	debug:    true,
		//	wantLog:  "cannot read file: EOF",
		//	wantCode: 1,
		//},
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
		tc := tc
		t.Run(testName, func(t *testing.T) {
			space, code, log := ABCIInfo(tc.err, tc.debug)
			if space != tc.wantSpace {
				t.Errorf("want %s space, got %s", tc.wantSpace, space)
			}
			if code != tc.wantCode {
				t.Errorf("want %d code, got %d", tc.wantCode, code)
			}
			if log != tc.wantLog {
				t.Errorf("want %q log, got %q", tc.wantLog, log)
			}
		})
	}
}

func TestABCIInfoStacktrace(t *testing.T) {
	cases := map[string]struct {
		err            error
		debug          bool
		wantStacktrace bool
		wantErrMsg     string
	}{
		"wrapped SDK error in debug mode provides stacktrace": {
			err:            Wrap(ErrUnauthorized, "wrapped"),
			debug:          true,
			wantStacktrace: true,
			wantErrMsg:     "unauthorized: wrapped",
		},
		"wrapped SDK error in non-debug mode does not have stacktrace": {
			err:            Wrap(ErrUnauthorized, "wrapped"),
			debug:          false,
			wantStacktrace: false,
			wantErrMsg:     "unauthorized: wrapped",
		},
		"wrapped stdlib error in debug mode provides stacktrace": {
			err:            Wrap(fmt.Errorf("stdlib"), "wrapped"),
			debug:          true,
			wantStacktrace: true,
			wantErrMsg:     "stdlib: wrapped",
		},
		"wrapped stdlib error in non-debug mode does not have stacktrace": {
			err:            Wrap(fmt.Errorf("stdlib"), "wrapped"),
			debug:          false,
			wantStacktrace: false,
			wantErrMsg:     "internal",
		},
	}

	const thisTestSrc = "github.com/cosmos/cosmos-sdk/types/errors.TestABCIInfoStacktrace"

	for testName, tc := range cases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			_, _, log := ABCIInfo(tc.err, tc.debug)
			if tc.wantStacktrace {
				if !strings.Contains(log, thisTestSrc) {
					t.Errorf("log does not contain this file stack trace: %s", log)
				}

				if !strings.Contains(log, tc.wantErrMsg) {
					t.Errorf("log does not contain expected error message: %s", log)
				}
			} else if log != tc.wantErrMsg {
				t.Fatalf("unexpected log message: %s", log)
			}
		})
	}
}

func TestABCIInfoHidesStacktrace(t *testing.T) {
	err := Wrap(ErrUnauthorized, "wrapped")
	_, _, log := ABCIInfo(err, false)

	if log != "unauthorized: wrapped" {
		t.Fatalf("unexpected message in non debug mode: %s", log)
	}
}

func TestRedact(t *testing.T) {
	cases := map[string]struct {
		err       error
		untouched bool  // if true we expect the same error after redact
		changed   error // if untouched == false, expect this error
	}{
		"panic looses message": {
			err:     Wrap(ErrPanic, "some secret stack trace"),
			changed: ErrPanic,
		},
		"sdk errors untouched": {
			err:       Wrap(ErrUnauthorized, "cannot drop db"),
			untouched: true,
		},
		"pass though custom errors with ABCI code": {
			err:       customErr{},
			untouched: true,
		},
		"redact stdlib error": {
			err:     fmt.Errorf("stdlib error"),
			changed: errInternal,
		},
	}

	for name, tc := range cases {
		spec := tc
		t.Run(name, func(t *testing.T) {
			redacted := Redact(spec.err)
			if spec.untouched {
				require.Equal(t, spec.err, redacted)
			} else {
				// see if we got the expected redact
				require.Equal(t, spec.changed, redacted)
				// make sure the ABCI code did not change
				require.Equal(t, abciCode(spec.err), abciCode(redacted))
			}
		})
	}
}

func TestABCIInfoSerializeErr(t *testing.T) {
	var (
		// Create errors with stacktrace for equal comparison.
		myErrDecode = Wrap(ErrTxDecode, "test")
		myErrAddr   = Wrap(ErrInvalidAddress, "tester")
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
			exp:   "tx parse error: test",
		},
		"second error": {
			src:   myErrAddr,
			debug: false,
			exp:   "invalid address: tester",
		},
		"single error with debug": {
			src:   myErrDecode,
			debug: true,
			exp:   fmt.Sprintf("%+v", myErrDecode),
		},
		"redact in default encoder": {
			src: myPanic,
			exp: "panic",
		},
		"do not redact in debug encoder": {
			src:   myPanic,
			debug: true,
			exp:   fmt.Sprintf("%+v", myPanic),
		},
	}
	for msg, spec := range specs {
		spec := spec
		t.Run(msg, func(t *testing.T) {
			_, _, log := ABCIInfo(spec.src, spec.debug)
			if exp, got := spec.exp, log; exp != got {
				t.Errorf("expected %v but got %v", exp, got)
			}
		})
	}
}

// customErr is a custom implementation of an error that provides an ABCICode
// method.
type customErr struct{}

func (customErr) Codespace() string { return "extern" }

func (customErr) ABCICode() uint32 { return 999 }

func (customErr) Error() string { return "custom" }
