package keyerror_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
)

func TestErrors(t *testing.T) {
	err := keyerror.NewErrKeyNotFound("test")
	require.True(t, keyerror.IsErrKeyNotFound(err))
	require.Equal(t, "Key test not found", err.Error())
	require.False(t, keyerror.IsErrKeyNotFound(errors.New("test")))
	require.False(t, keyerror.IsErrKeyNotFound(nil))

	err = keyerror.NewErrWrongPassword()
	require.True(t, keyerror.IsErrWrongPassword(err))
	require.Equal(t, "invalid account password", err.Error())
	require.False(t, keyerror.IsErrWrongPassword(errors.New("test")))
	require.False(t, keyerror.IsErrWrongPassword(nil))
}
