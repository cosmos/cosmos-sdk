package keys

import (
	"bufio"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_RunMnemonicCmdNormal(t *testing.T) {
	cmdBasic := mnemonicKeyCommand()
	err := runMnemonicCmd(cmdBasic, []string{})
	require.NoError(t, err)
}

func Test_RunMnemonicCmdUser(t *testing.T) {
	cmdUser := mnemonicKeyCommand()
	err := cmdUser.Flags().Set(flagUserEntropy, "1")
	assert.NoError(t, err)

	err = runMnemonicCmd(cmdUser, []string{})
	require.Error(t, err)
	require.Equal(t, "EOF", err.Error())

	// Try again
	cleanUp := client.OverrideStdin(bufio.NewReader(strings.NewReader("Hi!\n")))
	defer cleanUp()
	err = runMnemonicCmd(cmdUser, []string{})
	require.Error(t, err)
	require.Equal(t,
		"256-bits is 43 characters in Base-64, and 100 in Base-6. You entered 3, and probably want more",
		err.Error())

	// Now provide "good" entropy :)
	fakeEntropy := strings.Repeat(":)", 40) + "\ny\n" // entropy + accept count
	cleanUp2 := client.OverrideStdin(bufio.NewReader(strings.NewReader(fakeEntropy)))
	defer cleanUp2()
	err = runMnemonicCmd(cmdUser, []string{})
	require.NoError(t, err)

	// Now provide "good" entropy but no answer
	fakeEntropy = strings.Repeat(":)", 40) + "\n" // entropy + accept count
	cleanUp3 := client.OverrideStdin(bufio.NewReader(strings.NewReader(fakeEntropy)))
	defer cleanUp3()
	err = runMnemonicCmd(cmdUser, []string{})
	require.Error(t, err)

	// Now provide "good" entropy but say no
	fakeEntropy = strings.Repeat(":)", 40) + "\nn\n" // entropy + accept count
	cleanUp4 := client.OverrideStdin(bufio.NewReader(strings.NewReader(fakeEntropy)))
	defer cleanUp4()
	err = runMnemonicCmd(cmdUser, []string{})
	require.NoError(t, err)
}
