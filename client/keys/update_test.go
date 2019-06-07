package keys

import (
	"bufio"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/tests"
)

func Test_updateKeyCommand(t *testing.T) {
	cmd := updateKeyCommand()
	assert.NotNil(t, cmd)
	// No flags  or defaults to validate
}

func Test_runUpdateCmd(t *testing.T) {
	fakeKeyName1 := "runUpdateCmd_Key1"
	fakeKeyName2 := "runUpdateCmd_Key2"

	cmd := updateKeyCommand()

	// fails because it requests a password
	err := runUpdateCmd(cmd, []string{fakeKeyName1})
	assert.EqualError(t, err, "EOF")

	cleanUp := input.OverrideStdin(bufio.NewReader(strings.NewReader("pass1234\n")))
	defer cleanUp()

	// try again
	err = runUpdateCmd(cmd, []string{fakeKeyName1})
	assert.EqualError(t, err, "Key runUpdateCmd_Key1 not found")

	// Prepare a key base
	// Now add a temporary keybase
	kbHome, cleanUp1 := tests.NewTestCaseDir(t)
	defer cleanUp1()
	viper.Set(flags.FlagHome, kbHome)

	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", 0, 0)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", 0, 1)
	assert.NoError(t, err)

	// Try again now that we have keys
	cleanUp2 := input.OverrideStdin(bufio.NewReader(strings.NewReader("pass1234\nNew1234\nNew1234")))
	defer cleanUp2()

	// Incorrect key type
	err = runUpdateCmd(cmd, []string{fakeKeyName1})
	assert.EqualError(t, err, "locally stored key required. Received: keys.offlineInfo")

	// TODO: Check for other type types?

}
