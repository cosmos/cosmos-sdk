package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runExportCmd(t *testing.T) {
	runningUnattended := isRunningUnattended()
	config := sdk.NewDefaultConfig()
	exportKeyCommand := ExportKeyCommand(config)
	mockIn, _, _ := tests.ApplyMockIO(exportKeyCommand)

	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(flags.FlagHome, kbHome)

	// create a key
	kb, err := NewKeyringFromHomeFlag(mockIn, config)
	require.NoError(t, err)
	if !runningUnattended {
		defer func() {
			kb.Delete("keyname1", "", false)
		}()
	}

	if runningUnattended {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	_, err = kb.CreateAccount("keyname1", tests.TestMnemonic, "", "123456789", 0, 0)
	require.NoError(t, err)

	// Now enter password
	if runningUnattended {
		mockIn.Reset("123456789\n123456789\ntestpass1\n")
	} else {
		mockIn.Reset("123456789\n123456789\n")
	}
	require.NoError(t, runExportCmd(config)(exportKeyCommand, []string{"keyname1"}))
}
