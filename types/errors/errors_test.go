package errors

import (
	stdlib "errors"
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func TestCause(t *testing.T) {
	std := stdlib.New("this is a stdlib error")

	cases := map[string]struct {
		err  error
		root error
	}{
		"Errors are self-causing": {
			err:  ErrUnauthorized,
			root: ErrUnauthorized,
		},
		"Wrap reveals root cause": {
			err:  Wrap(ErrUnauthorized, "foo"),
			root: ErrUnauthorized,
		},
		"Cause works for stderr as root": {
			err:  Wrap(std, "Some helpful text"),
			root: std,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if got := errors.Cause(tc.err); got != tc.root {
				t.Fatal("unexpected result")
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	cases := map[string]struct {
		a      *Error
		b      error
		wantIs bool
	}{
		"instance of the same error": {
			a:      ErrUnauthorized,
			b:      ErrUnauthorized,
			wantIs: true,
		},
		"two different coded errors": {
			a:      ErrUnauthorized,
			b:      ErrOutOfGas,
			wantIs: false,
		},
		"successful comparison to a wrapped error": {
			a:      ErrUnauthorized,
			b:      errors.Wrap(ErrUnauthorized, "gone"),
			wantIs: true,
		},
		"unsuccessful comparison to a wrapped error": {
			a:      ErrUnauthorized,
			b:      errors.Wrap(ErrInsufficientFee, "too big"),
			wantIs: false,
		},
		"not equal to stdlib error": {
			a:      ErrUnauthorized,
			b:      fmt.Errorf("stdlib error"),
			wantIs: false,
		},
		"not equal to a wrapped stdlib error": {
			a:      ErrUnauthorized,
			b:      errors.Wrap(fmt.Errorf("stdlib error"), "wrapped"),
			wantIs: false,
		},
		"nil is nil": {
			a:      nil,
			b:      nil,
			wantIs: true,
		},
		"nil is any error nil": {
			a:      nil,
			b:      (*customError)(nil),
			wantIs: true,
		},
		"nil is not not-nil": {
			a:      nil,
			b:      ErrUnauthorized,
			wantIs: false,
		},
		"not-nil is not nil": {
			a:      ErrUnauthorized,
			b:      nil,
			wantIs: false,
		},
		// "multierr with the same error": {
		// 	a:      ErrUnauthorized,
		// 	b:      Append(ErrUnauthorized, ErrState),
		// 	wantIs: true,
		// },
		// "multierr with random order": {
		// 	a:      ErrUnauthorized,
		// 	b:      Append(ErrState, ErrUnauthorized),
		// 	wantIs: true,
		// },
		// "multierr with wrapped err": {
		// 	a:      ErrUnauthorized,
		// 	b:      Append(ErrState, Wrap(ErrUnauthorized, "test")),
		// 	wantIs: true,
		// },
		// "multierr with nil error": {
		// 	a:      ErrUnauthorized,
		// 	b:      Append(nil, nil),
		// 	wantIs: false,
		// },
		// "multierr with different error": {
		// 	a:      ErrUnauthorized,
		// 	b:      Append(ErrState, nil),
		// 	wantIs: false,
		// },
		// "multierr from nil": {
		// 	a:      nil,
		// 	b:      Append(ErrState, ErrUnauthorized),
		// 	wantIs: false,
		// },
		// "field error wrapper": {
		// 	a:      ErrEmpty,
		// 	b:      Field("name", ErrEmpty, "name is required"),
		// 	wantIs: true,
		// },
		// "nil field error wrapper": {
		// 	a:      nil,
		// 	b:      Field("name", nil, "name is required"),
		// 	wantIs: true,
		// },
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if got := tc.a.Is(tc.b); got != tc.wantIs {
				t.Fatalf("unexpected result - got:%v want: %v", got, tc.wantIs)
			}
		})
	}
}

type customError struct {
}

func (customError) Error() string {
	return "custom error"
}

func TestWrapEmpty(t *testing.T) {
	if err := Wrap(nil, "wrapping <nil>"); err != nil {
		t.Fatal(err)
	}
}
