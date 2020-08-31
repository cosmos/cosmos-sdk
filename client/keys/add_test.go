package keys

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runAddCmdBasic(t *testing.T) {
	runningUnattended := isRunningUnattended()
	cmd := AddKeyCommand()
	assert.NotNil(t, cmd)
	mockIn, _, _ := tests.ApplyMockIO(cmd)

	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(cli.OutputFlag, OutputFormatText)

	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	} else {
		mockIn.Reset("y\n")
		kb, err := keys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), kbHome, mockIn)
		require.NoError(t, err)
		defer func() {
			kb.Delete("keyname1", "", false)
			kb.Delete("keyname2", "", false)
		}()
	}
	assert.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	if runningUnattended {
		mockIn.Reset("testpass1\nN\n")
	} else {
		mockIn.Reset("N\n")
	}
	assert.Error(t, runAddCmd(cmd, []string{"keyname1"}))

	if runningUnattended {
		mockIn.Reset("testpass1\nN\n")
	} else {
		mockIn.Reset("y\n")
	}
	err := runAddCmd(cmd, []string{"keyname2"})
	assert.NoError(t, err)

	// In recovery mode
	cmd.SetArgs([]string{
		"keyname6",
		fmt.Sprintf("--%s=true", flagRecover),
	})

	// use valid mnemonic and complete recovery key generation successfully
	mockIn.Reset("decide praise business actor peasant farm drastic weather extend front hurt later song give verb rhythm worry fun pond reform school tumble august one\n")
	require.NoError(t, cmd.Execute())

	// use invalid mnemonic and fail recovery key generation
	mockIn.Reset("invalid mnemonic\n")
	require.Error(t, cmd.Execute())

	// In interactive mode
	cmd.SetArgs([]string{
		"keyname7",
		"-i",
		fmt.Sprintf("--%s=false", flagRecover),
	})

	const password = "password1!"

	// set password and complete interactive key generation successfully
	mockIn.Reset("\n" + password + "\n" + password + "\n")
	require.NoError(t, cmd.Execute())

	// passwords don't match and fail interactive key generation
	mockIn.Reset("\n" + password + "\n" + "fail" + "\n")
	require.Error(t, cmd.Execute())
}
