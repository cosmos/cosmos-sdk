package authz

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewGrant(t *testing.T) {
	a := NewGenericAuthorization("some-type")
	_, err := NewGrant(a, time.Unix(10, 0))
	require.NoError(t, err)
}
