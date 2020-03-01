package keys

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RunMnemonicCmdNormal(t *testing.T) {
	cmdBasic := MnemonicKeyCommand()
	require.NoError(t, runMnemonicCmd(cmdBasic, []string{}))
}

func Test_RunMnemonicCmdUser(t *testing.T) {
	cmdUser := MnemonicKeyCommand()
	err := cmdUser.Flags().Set(flagUserEntropy, "1")
	assert.NoError(t, err)

	err = runMnemonicCmd(cmdUser, []string{})
	require.Error(t, err)
	require.Equal(t, "EOF", err.Error())

	// Try again
	mockIn, _, _ := tests.ApplyMockIO(cmdUser)
	mockIn.Reset("Hi!\n")
	err = runMnemonicCmd(cmdUser, []string{})
	require.Error(t, err)
	require.Equal(t,
		"256-bits is 43 characters in Base-64, and 100 in Base-6. You entered 3, and probably want more",
		err.Error())

	// Now provide "good" entropy :)
	fakeEntropy := strings.Repeat(":)", 40) + "\ny\n" // entropy + accept count
	mockIn.Reset(fakeEntropy)
	require.NoError(t, runMnemonicCmd(cmdUser, []string{}))

	// Now provide "good" entropy but no answer
	fakeEntropy = strings.Repeat(":)", 40) + "\n" // entropy + accept count
	mockIn.Reset(fakeEntropy)
	require.Error(t, runMnemonicCmd(cmdUser, []string{}))

	// Now provide "good" entropy but say no
	fakeEntropy = strings.Repeat(":)", 40) + "\nn\n" // entropy + accept count
	mockIn.Reset(fakeEntropy)
	require.NoError(t, runMnemonicCmd(cmdUser, []string{}))
}
