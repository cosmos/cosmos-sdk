package keys

import (
	"testing"

	"github.com/99designs/keyring"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
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
	backends := keyring.AvailableBackends()
	runningOnServer := false

	if len(backends) == 2 && backends[1] == keyring.BackendType("file") {
		runningOnServer = true
	}

	viper.Set(flags.FlagLegacy, true)

	// fails because it requests a password
	assert.EqualError(t, runUpdateCmd(cmd, []string{fakeKeyName1}), "EOF")

	// try again
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	if runningOnServer {
		mockIn.Reset("testpass1\ny\ntestpass1\n")

	} else {
		mockIn.Reset("pass1234\n")
	}
	assert.EqualError(t, runUpdateCmd(cmd, []string{fakeKeyName1}), "Key runUpdateCmd_Key1 not found")

	// Prepare a key base
	// Now add a temporary keybase
	kbHome, cleanUp1 := tests.NewTestCaseDir(t)
	defer cleanUp1()
	viper.Set(flags.FlagHome, kbHome)

	kb := NewKeyringKeybase(mockIn)

	defer func() {
		kb.Delete("runUpdateCmd_Key1", "", false)
		kb.Delete("runUpdateCmd_Key2", "", false)
	}()
	if runningOnServer {
		mockIn.Reset("testpass1\ntestpass1\n")
	}
	_, err := kb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", 0, 0)
	assert.NoError(t, err)
	if runningOnServer {
		mockIn.Reset("testpass1\n")
	}
	_, err = kb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", 0, 1)
	assert.NoError(t, err)

	// Try again now that we have keys
	// Incorrect key type
	if runningOnServer {
		mockIn.Reset("testpass1\n")
	} else {
		mockIn.Reset("pass1234\nNew1234\nNew1234")
	}
	err = runUpdateCmd(cmd, []string{fakeKeyName1})
	assert.EqualError(t, err, "Key runUpdateCmd_Key1 not found")

	// TODO: Check for other type types?
}
