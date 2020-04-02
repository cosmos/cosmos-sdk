package errors

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
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
			wantLog:   "bar: foo: unauthorized",
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
			wantLog:   "internal error",
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
			wantLog:   "internal error",
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
			require.Equal(t, tc.wantSpace, space)
			require.Equal(t, tc.wantCode, code)
			require.Equal(t, tc.wantLog, log)
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
			wantErrMsg:     "wrapped: unauthorized",
		},
		"wrapped SDK error in non-debug mode does not have stacktrace": {
			err:            Wrap(ErrUnauthorized, "wrapped"),
			debug:          false,
			wantStacktrace: false,
			wantErrMsg:     "wrapped: unauthorized",
		},
		"wrapped stdlib error in debug mode provides stacktrace": {
			err:            Wrap(fmt.Errorf("stdlib"), "wrapped"),
			debug:          true,
			wantStacktrace: true,
			wantErrMsg:     "wrapped: stdlib",
		},
		"wrapped stdlib error in non-debug mode does not have stacktrace": {
			err:            Wrap(fmt.Errorf("stdlib"), "wrapped"),
			debug:          false,
			wantStacktrace: false,
			wantErrMsg:     "internal error",
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
	require.Equal(t, "wrapped: unauthorized", log)
}

func TestRedact(t *testing.T) {
	if err := Redact(ErrPanic); ErrPanic.Is(err) {
		t.Error("reduct must not pass through panic error")
	}
	if err := Redact(ErrUnauthorized); !ErrUnauthorized.Is(err) {
		t.Error("reduct should pass through SDK error")
	}

	var cerr customErr
	if err := Redact(cerr); err != cerr {
		t.Error("reduct should pass through ABCI code error")
	}

	serr := fmt.Errorf("stdlib error")
	if err := Redact(serr); err == serr {
		t.Error("reduct must not pass through a stdlib error")
	}

	require.Nil(t, Redact(nil))
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
		// "multi error default encoder": {
		// 	src: Append(myErrMsg, myErrAddr),
		// 	exp: Append(myErrMsg, myErrAddr).Error(),
		// },
		// "multi error default with internal": {
		// 	src: Append(myErrMsg, myPanic),
		// 	exp: "internal error",
		// },
		"redact in default encoder": {
			src: myPanic,
			exp: "internal error",
		},
		"do not redact in debug encoder": {
			src:   myPanic,
			debug: true,
			exp:   fmt.Sprintf("%+v", myPanic),
		},
		// 		"redact in multi error": {
		// 			src:   Append(myPanic, myErrMsg),
		// 			debug: false,
		// 			exp:   "internal error",
		// 		},
		// 		"no redact in multi error": {
		// 			src:   Append(myPanic, myErrMsg),
		// 			debug: true,
		// 			exp: `2 errors occurred:
		// 	* panic
		// 	* test: invalid message
		// `,
		// 		},
		// 		"wrapped multi error with redact": {
		// 			src:   Wrap(Append(myPanic, myErrMsg), "wrap"),
		// 			debug: false,
		// 			exp:   "internal error",
		// 		},
	}
	for msg, spec := range specs {
		spec := spec
		t.Run(msg, func(t *testing.T) {
			_, _, log := ABCIInfo(spec.src, spec.debug)
			require.Equal(t, spec.exp, log)
		})
	}
}

// customErr is a custom implementation of an error that provides an ABCICode
// method.
type customErr struct{}

func (customErr) Codespace() string { return "extern" }

func (customErr) ABCICode() uint32 { return 999 }

func (customErr) Error() string { return "custom" }

func TestResponseCheckDeliverTx(t *testing.T) {
	t.Parallel()
	require.Equal(t, abci.ResponseCheckTx{
		Codespace: "extern",
		Code:      999,
		Log:       "custom",
		GasWanted: int64(1),
		GasUsed:   int64(2),
	}, ResponseCheckTx(customErr{}, 1, 2))
	require.Equal(t, abci.ResponseDeliverTx{
		Codespace: "extern",
		Code:      999,
		Log:       "custom",
		GasWanted: int64(1),
		GasUsed:   int64(2),
	}, ResponseDeliverTx(customErr{}, 1, 2))
}

func TestQueryResult(t *testing.T) {
	t.Parallel()
	require.Equal(t, abci.ResponseQuery{
		Codespace: "extern",
		Code:      999,
		Log:       "custom",
	}, QueryResult(customErr{}))
}
