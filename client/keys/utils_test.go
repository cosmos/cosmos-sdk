package keys

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewKeyringFromDir(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	viper.Set(flags.FlagKeyringBackend, flags.KeyringBackendTest)
	_, err := NewKeyringFromDir(sdk.DefaultKeyringServiceName, filepath.Join(dir, "test"), nil)
	require.NoError(t, err)
	viper.Set(flags.FlagKeyringBackend, flags.KeyringBackendFile)
	buf := strings.NewReader("password\npassword\n")
	_, err = NewKeyringFromDir(sdk.DefaultKeyringServiceName, filepath.Join(dir, "test"), buf)
	require.NoError(t, err)
}
