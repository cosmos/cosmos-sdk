package cli_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/auth/client/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetSignCommand(t *testing.T) {
	clientCtx := client.Context{}

	encodingConfig := simappparams.MakeEncodingConfig()
	authtypes.RegisterCodec(encodingConfig.Amino)
	sdk.RegisterCodec(encodingConfig.Amino)
	encodingConfig.Amino.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)

	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	path := hd.CreateHDPath(118, 0, 0).String()
	kr, err := keyring.New(sdk.KeyringServiceName(), "test", dir, nil)
	require.NoError(t, err)

	var from = "test_sign"

	_, _, err = kr.NewMnemonic(from, keyring.English, path, hd.Secp256k1)
	require.NoError(t, err)
	require.NoError(t, kr.Delete(from))

	info, err := kr.NewAccount(from, tests.TestMnemonic, "", path, hd.Secp256k1)
	addr := info.GetAddress()
	require.NoError(t, err)

	//viper.Set(flags.FlagGenerateOnly, true)
	viper.Set(flags.FlagFrom, from)
	viper.Set(flags.FlagKeyringBackend, "test")
	viper.Set(flags.FlagHome, dir)

	clientCtx = clientCtx.WithTxGenerator(encodingConfig.TxGenerator).WithChainID("test").WithKeyring(kr).WithFrom(from)
	cmd := cli.GetSignCommand(clientCtx)

	txGen := clientCtx.TxGenerator

	// Build a test transaction
	fee := authtypes.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})

	stdTx := authtypes.NewStdTx([]sdk.Msg{authtypes.NewTestMsg(addr)}, fee, []authtypes.StdSignature{}, "foomemo")

	txJSONEncoded, err := txGen.TxJSONEncoder()(stdTx)
	require.NoError(t, err)

	txFile, cleanup := tests.WriteToNewTempFile(t, string(txJSONEncoded))
	txFileName := txFile.Name()
	// added for debugging purpose, (remove this lines after debugged)
	// fileData, err := ioutil.ReadFile(txFileName)
	// fmt.Println("fileData", string(fileData))
	t.Cleanup(cleanup)

	err = cmd.RunE(cmd, []string{txFileName})
	require.NoError(t, err)
}

func TestMultiSign(t *testing.T) {
	clientCtx := client.Context{}

	encodingConfig := simappparams.MakeEncodingConfig()
	authtypes.RegisterCodec(encodingConfig.Amino)
	sdk.RegisterCodec(encodingConfig.Amino)
	encodingConfig.Amino.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)

	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	path := hd.CreateHDPath(118, 0, 0).String()
	kr, err := keyring.New(sdk.KeyringServiceName(), "test", dir, nil)
	require.NoError(t, err)

	var from = "test_sign"

	_, _, err = kr.NewMnemonic(from, keyring.English, path, hd.Secp256k1)
	require.NoError(t, err)
	require.NoError(t, kr.Delete(from))

	info, err := kr.NewAccount(from, tests.TestMnemonic, "", path, hd.Secp256k1)
	addr := info.GetAddress()
	require.NoError(t, err)

	//viper.Set(flags.FlagGenerateOnly, true)
	viper.Set(flags.FlagFrom, from)
	viper.Set(flags.FlagKeyringBackend, "test")
	viper.Set(flags.FlagHome, dir)

	clientCtx = clientCtx.WithTxGenerator(encodingConfig.TxGenerator).WithChainID("test").WithKeyring(kr).WithFrom(from)
	cmd := cli.GetMultiSignCommand(clientCtx)

	txGen := clientCtx.TxGenerator

	// Build a test transaction
	fee := authtypes.NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})

	stdTx := authtypes.NewStdTx([]sdk.Msg{authtypes.NewTestMsg(addr)}, fee, []authtypes.StdSignature{}, "foomemo")

	txJSONEncoded, err := txGen.TxJSONEncoder()(stdTx)
	require.NoError(t, err)

	txFile, cleanup := tests.WriteToNewTempFile(t, string(txJSONEncoded))
	txFileName := txFile.Name()
	// added for debugging purpose, (remove this lines after debugged)
	// fileData, err := ioutil.ReadFile(txFileName)
	// fmt.Println("fileData", string(fileData))
	t.Cleanup(cleanup)

	err = cmd.RunE(cmd, []string{txFileName})
	require.NoError(t, err)
}
