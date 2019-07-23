//+build ledger test_ledger_mock

package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	bech32PrefixAccAddr = "terra"
	// bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	bech32PrefixAccPub = "terrapub"
	// bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	bech32PrefixValAddr = "terravaloper"
	// bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	bech32PrefixValPub = "terravaloperpub"
	// bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	bech32PrefixConsAddr = "terravalcons"
	// bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	bech32PrefixConsPub = "terravalconspub"
)

func Test_runAddCmdLedgerWithCustomCoinType(t *testing.T) {
	config := sdk.GetConfig()

	if config.GetBech32AccountAddrPrefix() != bech32PrefixAccAddr {
		config.SetCoinType(330)
		config.SetFullFundraiserPath("44'/330'/0'/0/0")
		config.SetBech32PrefixForAccount(bech32PrefixAccAddr, bech32PrefixAccPub)
		config.SetBech32PrefixForValidator(bech32PrefixValAddr, bech32PrefixValPub)
		config.SetBech32PrefixForConsensusNode(bech32PrefixConsAddr, bech32PrefixConsPub)
		config.Seal()
	}

	cmd := addKeyCommand()
	assert.NotNil(t, cmd)

	// Prepare a keybase
	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(flags.FlagUseLedger, true)

	/// Test Text
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	mockIn.Reset("test1234\ntest1234\n")
	assert.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	// Now check that it has been stored properly
	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	assert.NotNil(t, kb)
	key1, err := kb.Get("keyname1")
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	assert.Equal(t, "keyname1", key1.GetName())
	assert.Equal(t, keys.TypeLedger, key1.GetType())
	assert.Equal(t,
		"terrapub1addwnpepqvpg7r26nl2pvqqern00m6s9uaax3hauu2rzg8qpjzq9hy6xve7sw0d84m6",
		sdk.MustBech32ifyAccPub(key1.GetPubKey()))
}

func Test_runAddCmdLedger(t *testing.T) {
	cmd := addKeyCommand()
	assert.NotNil(t, cmd)

	// Prepare a keybase
	kbHome, kbCleanUp := tests.NewTestCaseDir(t)
	assert.NotNil(t, kbHome)
	defer kbCleanUp()
	viper.Set(flags.FlagHome, kbHome)
	viper.Set(flags.FlagUseLedger, true)

	/// Test Text
	viper.Set(cli.OutputFlag, OutputFormatText)
	// Now enter password
	mockIn, _, _ := tests.ApplyMockIO(cmd)
	mockIn.Reset("test1234\ntest1234\n")
	assert.NoError(t, runAddCmd(cmd, []string{"keyname1"}))

	// Now check that it has been stored properly
	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	assert.NotNil(t, kb)
	key1, err := kb.Get("keyname1")
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	assert.Equal(t, "keyname1", key1.GetName())
	assert.Equal(t, keys.TypeLedger, key1.GetType())
	assert.Equal(t,
		"cosmospub1addwnpepqd87l8xhcnrrtzxnkql7k55ph8fr9jarf4hn6udwukfprlalu8lgw0urza0",
		sdk.MustBech32ifyAccPub(key1.GetPubKey()))
}
