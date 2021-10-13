package expect

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpect(t *testing.T) {
	err := errors.New("some error message is here")
	tc := []struct {
		title     string
		err       error
		expectMsg string
		passes    bool
	}{
		{"handles nil-positive", nil, "", true},
		{"handles nil-negative", nil, "x", false},
		{"handles not-nil-positive", err, "error message", true},
		{"handles not-nil-negative-1", err, "", false},
		{"handles not-nil-negative-2", err, "not_there", false},
	}

	for _, tc := range tc {
		t.Run(tc.title, func(t *testing.T) {
			r := require.New(t)
			mockedT := new(MockT)
			ErrorContains(require.New(mockedT), tc.expectMsg, tc.err)
			r.NotEqual(tc.passes, mockedT.Failed)
		})
	}

}

type MockT struct {
	Failed bool
}

func (t *MockT) FailNow() {
	t.Failed = true
}

func (t *MockT) Errorf(format string, args ...interface{}) {
	_, _ = format, args
}
