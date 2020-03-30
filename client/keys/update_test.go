package keys

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/spf13/viper"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
)

func Test_runUpdateCmd(t *testing.T) {
	fakeKeyName1 := "runUpdateCmd_Key1"
	cmd := UpdateKeyCommand()
	require.NotNil(t, cmd)

	// Prepare a key base
	// Now add a temporary keybase
	kbHome, cleanUp1 := tests.NewTestCaseDir(t)
	t.Cleanup(cleanUp1)
	viper.Set(flags.FlagHome, kbHome)
	setUpLegacyKeybase(t, kbHome)

	// // Try again now that we have keys
	// // Incorrect key type
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	mockIn.Reset("pass1234\nNew1234\nNew1234")
	err := runUpdateCmd(cmd, []string{fakeKeyName1})
	require.EqualError(t, err, "locally stored key required. Received: keyring.offlineInfo")
}

func setUpLegacyKeybase(t *testing.T, dir string) {
	// Keys were created as follows:
	//	fakeKeyName1 := "runUpdateCmd_Key1"
	//	fakeKeyName2 := "runUpdateCmd_Key2"
	// _, err = fullKb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", "0", keyring.Secp256k1)
	// _, err = fullKb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", "1", keyring.Secp256k1)

	require.NoError(t, copy.Copy("testdata", dir))
	kb, err := NewLegacyKeyBaseFromDir(dir)
	require.NoError(t, err)
	keys, err := kb.List()
	require.NoError(t, err)

	require.Equal(t, 2, len(keys))
	require.NoError(t, kb.Close())
}
