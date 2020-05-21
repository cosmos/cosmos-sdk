package keys

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runImportCmd(t *testing.T) {
	runningUnattended := isRunningUnattended()
	importKeyCommand := ImportKeyCommand()
	mockIn, _, _ := tests.ApplyMockIO(importKeyCommand)

	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(flags.FlagHome, kbHome)

	if !runningUnattended {
		kb, err := keys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), mockIn)
		require.NoError(t, err)
		defer func() {
			kb.Delete("keyname1", "", false)
		}()
	}

	keyfile := filepath.Join(kbHome, "key.asc")
	armoredKey := `-----BEGIN TENDERMINT PRIVATE KEY-----
salt: A790BB721D1C094260EA84F5E5B72289
kdf: bcrypt

HbP+c6JmeJy9JXe2rbbF1QtCX1gLqGcDQPBXiCtFvP7/8wTZtVOPj8vREzhZ9ElO
3P7YnrzPQThG0Q+ZnRSbl9MAS8uFAM4mqm5r/Ys=
=f3l4
-----END TENDERMINT PRIVATE KEY-----
`
	require.NoError(t, ioutil.WriteFile(keyfile, []byte(armoredKey), 0600))

	// Now enter password
	if runningUnattended {
		mockIn.Reset("123456789\n12345678\n12345678\n")
	} else {
		mockIn.Reset("123456789\n")
	}
	require.NoError(t, runImportCmd(importKeyCommand, []string{"keyname1", keyfile}))
}
