package flags_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

func TestKeyringBackend(t *testing.T) {
	b := flags.KeyringBackend{Backend: "os"}
	require.Equal(t, "string", b.Type())
	require.Error(t, b.Set("unknown"))
	require.Equal(t, "os", b.String())
	require.NoError(t, b.Set("file"))
	require.Equal(t, "file", b.String())
	require.NoError(t, b.Set(""))
	require.Equal(t, "os", b.String())
}
