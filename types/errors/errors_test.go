package errors

import (
	stdlib "errors"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type errorsTestSuite struct {
	suite.Suite
}

func TestErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(errorsTestSuite))
}

func (s *errorsTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *errorsTestSuite) TestCause() {
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
		s.Require().Equal(tc.root, errors.Cause(tc.err), testName)
	}
}

func (s *errorsTestSuite) TestErrorIs() {
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
		s.Require().Equal(tc.wantIs, tc.a.Is(tc.b), testName)
	}
}

type customError struct {
}

func (customError) Error() string {
	return "custom error"
}

func (s *errorsTestSuite) TestWrapEmpty() {
	s.Require().Nil(Wrap(nil, "wrapping <nil>"))
}

func (s *errorsTestSuite) TestWrappedIs() {
	require := s.Require()
	err := Wrap(ErrTxTooLarge, "context")
	require.True(stdlib.Is(err, ErrTxTooLarge))

	err = Wrap(err, "more context")
	require.True(stdlib.Is(err, ErrTxTooLarge))

	err = Wrap(err, "even more context")
	require.True(stdlib.Is(err, ErrTxTooLarge))

	err = Wrap(ErrInsufficientFee, "...")
	require.False(stdlib.Is(err, ErrTxTooLarge))

	errs := stdlib.New("other")
	require.True(stdlib.Is(errs, errs))

	errw := &wrappedError{"msg", errs}
	require.True(errw.Is(errw), "should match itself")
}

func (s *errorsTestSuite) TestWrappedIsMultiple() {
	var errTest = errors.New("test error")
	var errTest2 = errors.New("test error 2")
	err := Wrap(errTest2, Wrap(errTest, "some random description").Error())
	s.Require().True(stdlib.Is(err, errTest2))
}

func (s *errorsTestSuite) TestWrappedIsFail() {
	var errTest = errors.New("test error")
	var errTest2 = errors.New("test error 2")
	err := Wrap(errTest2, Wrap(errTest, "some random description").Error())
	s.Require().False(stdlib.Is(err, errTest))
}

func (s *errorsTestSuite) TestWrappedUnwrap() {
	var errTest = errors.New("test error")
	err := Wrap(errTest, "some random description")
	s.Require().Equal(errTest, stdlib.Unwrap(err))
}

func (s *errorsTestSuite) TestWrappedUnwrapMultiple() {
	var errTest = errors.New("test error")
	var errTest2 = errors.New("test error 2")
	err := Wrap(errTest2, Wrap(errTest, "some random description").Error())
	s.Require().Equal(errTest2, stdlib.Unwrap(err))
}

func (s *errorsTestSuite) TestWrappedUnwrapFail() {
	var errTest = errors.New("test error")
	var errTest2 = errors.New("test error 2")
	err := Wrap(errTest2, Wrap(errTest, "some random description").Error())
	s.Require().NotEqual(errTest, stdlib.Unwrap(err))
}

func (s *errorsTestSuite) TestABCIError() {
	s.Require().Equal("custom: tx parse error", ABCIError(RootCodespace, 2, "custom").Error())
	s.Require().Equal("custom: unknown", ABCIError("unknown", 1, "custom").Error())
}

func ExampleWrap() {
	err1 := Wrap(ErrInsufficientFunds, "90 is smaller than 100")
	err2 := errors.Wrap(ErrInsufficientFunds, "90 is smaller than 100")
	fmt.Println(err1.Error())
	fmt.Println(err2.Error())
	// Output:
	// 90 is smaller than 100: insufficient funds
	// 90 is smaller than 100: insufficient funds
}

func ExampleWrapf() {
	err1 := Wrap(ErrInsufficientFunds, "90 is smaller than 100")
	err2 := errors.Wrap(ErrInsufficientFunds, "90 is smaller than 100")
	fmt.Println(err1.Error())
	fmt.Println(err2.Error())
	// Output:
	// 90 is smaller than 100: insufficient funds
	// 90 is smaller than 100: insufficient funds
}
