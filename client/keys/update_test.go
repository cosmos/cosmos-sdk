package keys

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func Test_updateKeyCommand(t *testing.T) {
	require.NotNil(t, UpdateKeyCommand())
}

func Test_runUpdateCmd(t *testing.T) {
	fakeKeyName1 := "runUpdateCmd_Key1"
	fakeKeyName2 := "runUpdateCmd_Key2"
	cmd := UpdateKeyCommand()

	// Prepare a key base
	// Now add a temporary keybase
	kbHome, cleanUp1 := tests.NewTestCaseDir(t)
	t.Cleanup(cleanUp1)
	viper.Set(flags.FlagHome, kbHome)

	kb, err := NewLegacyKeyBaseFromDir(kbHome)

	fullKb, ok := kb.(keyring.Keybase)
	require.True(t, ok)

	require.NoError(t, err)
	_, err = fullKb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", "0", keyring.Secp256k1)
	require.NoError(t, err)
	_, err = fullKb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", "1", keyring.Secp256k1)
	require.NoError(t, err)
	require.NoError(t, kb.Close())

	// Try again now that we have keys
	// Incorrect key type
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	mockIn.Reset("pass1234\nNew1234\nNew1234")
	err = runUpdateCmd(cmd, []string{fakeKeyName1})
	require.EqualError(t, err, "locally stored key required. Received: keyring.offlineInfo")

	// TODO: Check for other type types?
}
