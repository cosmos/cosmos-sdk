package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MultiErrorTestSuite struct {
	suite.Suite

	err1 error
	err2 error
	err3 error
	err4 error
}

func TestMultiErrorTestSuite(t *testing.T) {
	suite.Run(t, new(MultiErrorTestSuite))
}

func (s *MultiErrorTestSuite) SetupTest() {
	s.err1 = errors.New("expected error one")
	s.err2 = errors.New("expected error two")
	s.err3 = errors.New("expected error three")
	s.err3 = errors.New("expected error four")
}

func (s *MultiErrorTestSuite) TestFlattenErrors() {
	tests := []struct {
		name     string
		input    []error
		expected error
	}{
		{
			name:     "none in nil out",
			input:    []error{},
			expected: nil,
		},
		{
			name:     "nil in nil out",
			input:    []error{nil},
			expected: nil,
		},
		{
			name:     "nils in nil out",
			input:    []error{nil, nil, nil},
			expected: nil,
		},
		{
			name:     "one in same out",
			input:    []error{s.err1},
			expected: s.err1,
		},
		{
			name:     "nils and one in that one out",
			input:    []error{nil, s.err2, nil},
			expected: s.err2,
		},
		{
			name:     "two in multi out with both",
			input:    []error{s.err1, s.err2},
			expected: &MultiError{errs: []error{s.err1, s.err2}},
		},
		{
			name:     "two and nils in multi out with both",
			input:    []error{nil, s.err1, nil, s.err2, nil},
			expected: &MultiError{errs: []error{s.err1, s.err2}},
		},
		{
			name:     "lots in multi out",
			input:    []error{s.err1, s.err2, s.err3, s.err2, s.err1},
			expected: &MultiError{errs: []error{s.err1, s.err2, s.err3, s.err2, s.err1}},
		},
		{
			name:     "multi and non in one multi out with all",
			input:    []error{&MultiError{errs: []error{s.err1, s.err2}}, s.err3},
			expected: &MultiError{errs: []error{s.err1, s.err2, s.err3}},
		},
		{
			name:     "non and multi in one multi out with all",
			input:    []error{s.err1, &MultiError{errs: []error{s.err2, s.err3}}},
			expected: &MultiError{errs: []error{s.err1, s.err2, s.err3}},
		},
		{
			name:     "two multi in one multi out with all",
			input:    []error{&MultiError{errs: []error{s.err1, s.err2}}, &MultiError{errs: []error{s.err3, s.err4}}},
			expected: &MultiError{errs: []error{s.err1, s.err2, s.err3, s.err4}},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := FlattenErrors(tc.input...)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func (s *MultiErrorTestSuite) TestGetErrors() {
	tests := []struct {
		name     string
		multi    MultiError
		expected []error
	}{
		{
			name:     "two",
			multi:    MultiError{errs: []error{s.err3, s.err1}},
			expected: []error{s.err3, s.err1},
		},
		{
			name:     "three",
			multi:    MultiError{errs: []error{s.err3, s.err1, s.err2}},
			expected: []error{s.err3, s.err1, s.err2},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			// Make sure it's getting what's expected.
			actual1 := tc.multi.GetErrors()
			require.NotSame(t, tc.expected, actual1)
			require.Equal(t, tc.expected, actual1)
			// Make sure that changing what was given back doesn't alter the original.
			actual1[0] = errors.New("unexpected error")
			actual2 := tc.multi.GetErrors()
			require.NotEqual(t, actual1, actual2)
			require.Equal(t, tc.expected, actual2)
		})
	}
}

func (s *MultiErrorTestSuite) TestLen() {
	tests := []struct {
		name     string
		multi    MultiError
		expected int
	}{
		{
			name:     "two",
			multi:    MultiError{errs: []error{s.err3, s.err1}},
			expected: 2,
		},
		{
			name:     "three",
			multi:    MultiError{errs: []error{s.err3, s.err1, s.err2}},
			expected: 3,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := tc.multi.Len()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func (s *MultiErrorTestSuite) TestErrorAndString() {
	tests := []struct {
		name     string
		multi    MultiError
		expected string
	}{
		{
			name:     "two",
			multi:    MultiError{errs: []error{s.err1, s.err2}},
			expected: fmt.Sprintf("2 errors: 1: %s, 2: %s", s.err1, s.err2),
		},
		{
			name:     "three",
			multi:    MultiError{errs: []error{s.err1, s.err2, s.err3}},
			expected: fmt.Sprintf("3 errors: 1: %s, 2: %s, 3: %s", s.err1, s.err2, s.err3),
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name+" Error", func(t *testing.T) {
			actual := tc.multi.Error()
			require.Equal(t, tc.expected, actual)
		})
		s.T().Run(tc.name+" String", func(t *testing.T) {
			actual := tc.multi.String()
			require.Equal(t, tc.expected, actual)
		})
	}
}
