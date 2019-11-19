//+build ledger test_ledger_mock

package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runAddCmdLedgerWithCustomCoinType(t *testing.T) {
	runningUnattended := isRunningUnattended()
	config := sdk.GetConfig()

	bech32PrefixAccAddr := "terra"
	bech32PrefixAccPub := "terrapub"
	bech32PrefixValAddr := "terravaloper"
	bech32PrefixValPub := "terravaloperpub"
	bech32PrefixConsAddr := "terravalcons"
	bech32PrefixConsPub := "terravalconspub"

	config.SetCoinType(330)
	config.SetFullFundraiserPath("44'/330'/0'/0/0")
	config.SetBech32PrefixForAccount(bech32PrefixAccAddr, bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(bech32PrefixValAddr, bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(bech32PrefixConsAddr, bech32PrefixConsPub)

	cmd := addKeyCommand()
	require.NotNil(t, cmd)

	// Prepare a keybase
	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	require.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(flags.FlagUseLedger, true)

	/// Test Text
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	mockIn.Reset("test1234\ntest1234\n")
	require.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	// Now check that it has been stored properly
	kb, err := NewKeyringFromHomeFlag(mockIn)
	require.NoError(t, err)
	require.NotNil(t, kb)
	defer func() {
		kb.Delete("keyname1", "", false)
	}()
	mockIn.Reset("test1234\n")
	if runningUnattended {
		mockIn.Reset("test1234\ntest1234\n")
	}
	key1, err := kb.Get("keyname1")
	require.NoError(t, err)
	require.NotNil(t, key1)

	require.Equal(t, "keyname1", key1.GetName())
	require.Equal(t, keys.TypeLedger, key1.GetType())
	require.Equal(t,
		"terrapub1addwnpepqvpg7r26nl2pvqqern00m6s9uaax3hauu2rzg8qpjzq9hy6xve7sw0d84m6",
		sdk.MustBech32ifyAccPub(key1.GetPubKey()))

	config.SetCoinType(118)
	config.SetFullFundraiserPath("44'/118'/0'/0/0")
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
}

func Test_runAddCmdLedger(t *testing.T) {
	runningUnattended := isRunningUnattended()
	cmd := addKeyCommand()
	require.NotNil(t, cmd)
	mockIn, _, _ := tests.ApplyMockIO(cmd)

	// Prepare a keybase
	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	require.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(flags.FlagUseLedger, true)

	/// Test Text
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	mockIn.Reset("test1234\ntest1234\n")
	require.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	// Now check that it has been stored properly
	kb, err := NewKeyringFromHomeFlag(mockIn)
	require.NoError(t, err)
	require.NotNil(t, kb)
	defer func() {
		kb.Delete("keyname1", "", false)
	}()
	mockIn.Reset("test1234\n")
	if runningUnattended {
		mockIn.Reset("test1234\ntest1234\n")
	}
	key1, err := kb.Get("keyname1")
	require.NoError(t, err)
	require.NotNil(t, key1)

	require.Equal(t, "keyname1", key1.GetName())
	require.Equal(t, keys.TypeLedger, key1.GetType())
	require.Equal(t,
		"cosmospub1addwnpepqd87l8xhcnrrtzxnkql7k55ph8fr9jarf4hn6udwukfprlalu8lgw0urza0",
		sdk.MustBech32ifyAccPub(key1.GetPubKey()))
}
