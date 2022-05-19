package authz

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TODO: remove and use: robert/expect-error
func expecError(r *require.Assertions, expected string, received error) {
	if expected == "" {
		r.NoError(received)
	} else {
		r.Error(received)
		r.Contains(received.Error(), expected)
	}
}

func TestNewGrant(t *testing.T) {
	a := NewGenericAuthorization("some-type")
	tcs := []struct {
		title     string
		a         Authorization
		blockTime time.Time
		expire    *time.Time
		err       string
	}{
		{"wrong expire time (1)", a, time.Unix(10, 0), unixTime(8, 0), "expiration must be after"},
		{"wrong expire time (2)", a, time.Unix(10, 0), unixTime(10, 0), "expiration must be after"},
		{"good expire time (1)", a, time.Unix(10, 0), unixTime(10, 1), ""},
		{"good expire time (2)", a, time.Unix(10, 0), unixTime(11, 0), ""},
		{"good expire time (nil)", a, time.Unix(10, 0), nil, ""},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			_, err := NewGrant(tc.blockTime, tc.a, tc.expire)
			expecError(require.New(t), tc.err, err)
		})
	}
}

func unixTime(s, ns int64) *time.Time {
	t := time.Unix(s, ns)
	return &t
}
