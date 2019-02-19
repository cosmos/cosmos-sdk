package keys

import (
	"bufio"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/stretchr/testify/assert"
)

func Test_runAddCmdBasic(t *testing.T) {
	cmd := addKeyCommand()
	assert.NotNil(t, cmd)

	// Prepare a keybase
	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(cli.HomeFlag, kbHome)

	/// Test Text
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	cleanUp1 := client.OverrideStdin(bufio.NewReader(strings.NewReader("test1234\ntest1234\n")))
	defer cleanUp1()
	err := runAddCmd(cmd, []string{"keyname1"})
	assert.NoError(t, err)

	/// Test Text - Replace? >> FAIL
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	cleanUp2 := client.OverrideStdin(bufio.NewReader(strings.NewReader("test1234\ntest1234\n")))
	defer cleanUp2()
	err = runAddCmd(cmd, []string{"keyname1"})
	assert.Error(t, err)

	/// Test Text - Replace? Answer >> PASS
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	cleanUp3 := client.OverrideStdin(bufio.NewReader(strings.NewReader("y\ntest1234\ntest1234\n")))
	defer cleanUp3()
	err = runAddCmd(cmd, []string{"keyname1"})
	assert.NoError(t, err)

	// Check JSON
	viper.Set(cli.OutputFlag, OutputFormatJSON)
	// Now enter password
	cleanUp4 := client.OverrideStdin(bufio.NewReader(strings.NewReader("test1234\ntest1234\n")))
	defer cleanUp4()
	err = runAddCmd(cmd, []string{"keyname2"})
	assert.NoError(t, err)
}
