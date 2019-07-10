package keys

import (
	"testing"

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

	// create a key
	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	_, err = kb.CreateAccount("keyname1", tests.TestMnemonic, "", "123456789", 0, 0)
	assert.NoError(t, err)

	mockIn, _, _ := tests.ApplyMockIO(exportKeyCommand)
	mockIn.Reset("123456789\n123456789\n")
	// Now enter password
	assert.NoError(t, runExportCmd(exportKeyCommand, []string{"keyname1"}))
}
