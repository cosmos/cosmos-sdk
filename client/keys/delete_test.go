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

func Test_runDeleteCmd(t *testing.T) {
	deleteKeyCommand := DeleteKeyCommand()
	mockIn, _, _ := tests.ApplyMockIO(deleteKeyCommand)

	yesF, _ := deleteKeyCommand.Flags().GetBool(flagYes)
	forceF, _ := deleteKeyCommand.Flags().GetBool(flagForce)

	require.False(t, yesF)
	require.False(t, forceF)

	fakeKeyName1 := "runDeleteCmd_Key1"
	fakeKeyName2 := "runDeleteCmd_Key2"
	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	t.Cleanup(cleanUp)
	viper.Set(flags.FlagHome, kbHome)

	// Now
	path := sdk.GetConfig().GetFullFundraiserPath()
	backend := viper.GetString(flags.FlagKeyringBackend)
	kb, err := keyring.New(sdk.KeyringServiceName(), backend, kbHome, mockIn)
	require.NoError(t, err)
	_, err = kb.NewAccount(fakeKeyName1, tests.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)
	_, _, err = kb.NewMnemonic(fakeKeyName2, keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	require.NoError(t, err)

	err = runDeleteCmd(deleteKeyCommand, []string{"blah"})
	require.Error(t, err)
	require.Equal(t, "The specified item could not be found in the keyring", err.Error())

	// User confirmation missing
	err = runDeleteCmd(deleteKeyCommand, []string{fakeKeyName1})
	require.Error(t, err)
	require.Equal(t, "EOF", err.Error())

	_, err = kb.Key(fakeKeyName1)
	require.NoError(t, err)

	// Now there is a confirmation
	viper.Set(flagYes, true)
	require.NoError(t, runDeleteCmd(deleteKeyCommand, []string{fakeKeyName1}))

	_, err = kb.Key(fakeKeyName1)
	require.Error(t, err) // Key1 is gone

	viper.Set(flagYes, true)
	_, err = kb.Key(fakeKeyName2)
	require.NoError(t, err)
	err = runDeleteCmd(deleteKeyCommand, []string{fakeKeyName2})
	require.NoError(t, err)
	_, err = kb.Key(fakeKeyName2)
	require.Error(t, err) // Key2 is gone
}
