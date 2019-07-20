package keys

import (
	"testing"

	"github.com/99designs/keyring"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
)

func Test_runExportCmd(t *testing.T) {
	exportKeyCommand := exportKeyCommand()

	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(flags.FlagHome, kbHome)

	mockIn, _, _ := tests.ApplyMockIO(exportKeyCommand)

	backends := keyring.AvailableBackends()

	runningOnServer := false

	if len(backends) == 2 && backends[1] == keyring.BackendType("file") {
		runningOnServer = true
	}

	kb := NewKeyringKeybase(mockIn)

	if !runningOnServer {
		defer func() {
			kb.Delete("keyname1", "", false)
		}()
	}

	if runningOnServer {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	_, err := kb.CreateAccount("keyname1", tests.TestMnemonic, "", "123456789", 0, 0)
	assert.NoError(t, err)

	if runningOnServer {
		mockIn.Reset("123456789\n123456789\ntestpass1\ntestpass1\n")
	} else {
		mockIn.Reset("123456789\n123456789\n")

	}
	// Now enter password
	assert.NoError(t, runExportCmd(exportKeyCommand, []string{"keyname1"}))
}
