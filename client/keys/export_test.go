package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runExportCmd(t *testing.T) {
	exportKeyCommand := ExportKeyCommand()
	mockIn, _, _ := tests.ApplyMockIO(exportKeyCommand)

	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	t.Cleanup(cleanUp)
	viper.Set(flags.FlagHome, kbHome)

	// create a key
	kb, err := keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), kbHome, mockIn)
	require.NoError(t, err)
	t.Cleanup(func() {
		kb.Delete("keyname1") // nolint:errcheck
	})

	path := sdk.GetConfig().GetFullFundraiserPath()
	_, err = kb.NewAccount("keyname1", tests.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)

	// Now enter password
	mockIn.Reset("123456789\n123456789\n")
	require.NoError(t, runExportCmd(exportKeyCommand, []string{"keyname1"}))
}
