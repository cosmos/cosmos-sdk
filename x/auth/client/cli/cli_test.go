package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	codec2 "github.com/cosmos/cosmos-sdk/codec"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtest "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestCLIValidateSignatures() {
	val := s.network.Validators[0]
	res, err := bankcli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
			sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	var tx types.StdTx
	err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(res, &tx)
	s.Require().NoError(err)

	// write  unsigned tx to file
	unsignedTx, cleanup := testutil.WriteToNewTempFile(s.T(), string(res))
	defer cleanup()

	res, err = authtest.TxSignExec(val.ClientCtx, val.Address, unsignedTx.Name())
	s.Require().NoError(err)

	var signedTx types.StdTx
	err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(res, &signedTx)
	s.Require().NoError(err)

	signedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), string(res))
	defer cleanup()

	res, err = authtest.TxValidateSignaturesExec(val.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	signedTx.Memo = "MODIFIED STD TX"
	bz, err := val.ClientCtx.JSONMarshaler.MarshalJSON(signedTx)
	s.Require().NoError(err)

	modifiedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), string(bz))
	defer cleanup()

	res, err = authtest.TxValidateSignaturesExec(val.ClientCtx, modifiedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
}

func (s *IntegrationTestSuite) TestCLISignBatch() {
	val := s.network.Validators[0]
	generatedStd, err := bankcli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
			sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)

	s.Require().NoError(err)

	// Write the output to disk
	filename, cleanup1 := testutil.WriteToNewTempFile(s.T(), strings.Repeat(string(generatedStd), 3))
	defer cleanup1()

	// sign-batch file - offline is set but account-number and sequence are not
	val.ClientCtx.HomeDir = strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)
	res, err := authtest.TxSignBatchExec(val.ClientCtx, val.Address, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// sign-batch file
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(string(res), "\n"), "\n")))

	// sign-batch file
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().Equal(3, len(strings.Split(strings.Trim(string(res), "\n"), "\n")))

	// Sign batch malformed tx file.
	malformedFile, cleanup2 := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf("%smalformed", generatedStd))
	defer cleanup2()

	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID))
	s.Require().EqualError(err, "cannot parse disfix JSON wrapper: invalid character 'm' looking for beginning of value")

	// Sign batch malformed tx file signature only.
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().EqualError(err, "cannot parse disfix JSON wrapper: invalid character 'm' looking for beginning of value")
}

func (s *IntegrationTestSuite) TestCLISendGenerateSignAndBroadcast() {
	val1 := s.network.Validators[0]

	account, _, err := val1.ClientCtx.Keyring.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	sendTokens := sdk.TokensFromConsensusPower(10)

	normalGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		account.GetAddress(),
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, sendTokens),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	normalGeneratedStdTx := cli.UnmarshalStdTx(s.T(), val1.ClientCtx.JSONMarshaler, string(normalGeneratedTx))
	s.Require().Equal(normalGeneratedStdTx.Fee.Gas, uint64(flags.DefaultGasLimit))
	s.Require().Equal(len(normalGeneratedStdTx.Msgs), 1)
	s.Require().Equal(0, len(normalGeneratedStdTx.GetSignatures()))

	// Test generate sendTx with --gas=$amount
	limitedGasGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		account.GetAddress(),
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, sendTokens),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", 100),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	limitedGasStdTx := cli.UnmarshalStdTx(s.T(), val1.ClientCtx.JSONMarshaler, string(limitedGasGeneratedTx))
	s.Require().Equal(limitedGasStdTx.Fee.Gas, uint64(100))
	s.Require().Equal(len(limitedGasStdTx.Msgs), 1)
	s.Require().Equal(0, len(limitedGasStdTx.GetSignatures()))

	// Test generate sendTx, estimate gas
	finalGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		account.GetAddress(),
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, sendTokens),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	finalStdTx := cli.UnmarshalStdTx(s.T(), val1.ClientCtx.JSONMarshaler, string(finalGeneratedTx))
	s.Require().Equal(uint64(flags.DefaultGasLimit), finalStdTx.Fee.Gas)
	s.Require().Equal(len(finalStdTx.Msgs), 1)

	// Write the output to disk
	unsignedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), string(finalGeneratedTx))
	defer cleanup()

	// Test validate-signatures
	res, err := authtest.TxValidateSignaturesExec(val1.ClientCtx, unsignedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
	s.Require().True(strings.Contains(string(res), fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n\n", val1.Address.String())))

	// Test sign

	// Does not work in offline mode
	res, err = authtest.TxSignExec(val1.ClientCtx, val1.Address, unsignedTxFile.Name(), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// But works offline if we set account number and sequence
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	res, err = authtest.TxSignExec(val1.ClientCtx, val1.Address, unsignedTxFile.Name(), "--offline", "--account-number", "1", "--sequence", "1")
	s.Require().NoError(err)

	// Sign transaction
	signedTx, err := authtest.TxSignExec(val1.ClientCtx, val1.Address, unsignedTxFile.Name())
	s.Require().NoError(err)

	signedFinalTx := cli.UnmarshalStdTx(s.T(), val1.ClientCtx.JSONMarshaler, string(signedTx))
	s.Require().Equal(len(signedFinalTx.Msgs), 1)
	s.Require().Equal(1, len(signedFinalTx.GetSignatures()))
	s.Require().Equal(val1.Address.String(), signedFinalTx.GetSigners()[0].String())

	// Write the output to disk
	signedTxFile, cleanup2 := testutil.WriteToNewTempFile(s.T(), string(signedTx))
	defer cleanup2()

	// Validate Signature
	res, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)
	s.Require().True(strings.Contains(string(res), "[OK]"))

	// Ensure foo has right amount of funds
	startTokens := sdk.TokensFromConsensusPower(400)
	resp, err := bankcli.QueryBalancesExec(val1.ClientCtx, val1.Address)
	s.Require().NoError(err)

	var coins sdk.Coins
	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &coins)
	s.Require().NoError(err)
	s.Require().Equal(startTokens, coins.AmountOf(cli.Denom))

	// Test broadcast

	// Does not work in offline mode
	res, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name(), "--offline")
	s.Require().EqualError(err, "cannot broadcast tx during offline mode")

	err = s.network.WaitForNBlocks(1, 10*time.Second)
	s.Require().NoError(err)

	// Broadcast correct transaction.
	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	res, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	err = s.network.WaitForNBlocks(1, 10*time.Second)
	s.Require().NoError(err)

	// Ensure destiny account state
	resp, err = bankcli.QueryBalancesExec(val1.ClientCtx, account.GetAddress())
	s.Require().NoError(err)

	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &coins)
	s.Require().NoError(err)
	s.Require().Equal(sendTokens, coins.AmountOf(cli.Denom))

	// Ensure origin account state
	resp, err = bankcli.QueryBalancesExec(val1.ClientCtx, val1.Address)
	s.Require().NoError(err)

	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &coins)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(389999990), coins.AmountOf(cli.Denom))
}

func (s *IntegrationTestSuite) TestCLIMultisignInsufficientCosigners() {
	s.T().SkipNow() // TODO check encoding.
	val1 := s.network.Validators[0]

	codec := codec2.New()
	sdk.RegisterCodec(codec)
	banktypes.RegisterCodec(codec)
	val1.ClientCtx.Codec = codec

	// Generate 2 accounts and a multisig.
	account1, _, err := val1.ClientCtx.Keyring.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	account2, _, err := val1.ClientCtx.Keyring.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	multi := multisig.NewPubKeyMultisigThreshold(2, []tmcrypto.PubKey{account1.GetPubKey(), account2.GetPubKey()})
	multisigInfo, err := val1.ClientCtx.Keyring.SaveMultisig("multi", multi)
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	_, err = bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		multisigInfo.GetAddress(),
		sdk.NewCoins(
			sdk.NewInt64Coin(cli.Denom, 10),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNBlocks(1, 10*time.Second)
	s.Require().NoError(err)

	// Generate multisig transaction.
	multiGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		multisigInfo.GetAddress(),
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(cli.Denom, 5),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), string(multiGeneratedTx))
	defer cleanup()

	// Multisign, sign with one signature
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File, cleanup2 := testutil.WriteToNewTempFile(s.T(), string(account1Signature))
	defer cleanup2()

	multiSigWith1Signature, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name())
	s.Require().NoError(err)

	// Save tx to file
	multiSigWith1SignatureFile, cleanup3 := testutil.WriteToNewTempFile(s.T(), string(multiSigWith1Signature))
	defer cleanup3()

	exec, err := authtest.TxValidateSignaturesExec(val1.ClientCtx, multiSigWith1SignatureFile.Name())
	s.Require().NoError(err)

	fmt.Printf("%s", exec)
}

func (s *IntegrationTestSuite) TestCLIEncode() {
	val1 := s.network.Validators[0]

	sendTokens := sdk.TokensFromConsensusPower(10)

	normalGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		val1.Address,
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, sendTokens),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly), "--memo", "deadbeef",
	)
	s.Require().NoError(err)

	// Save tx to file
	savedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), string(normalGeneratedTx))
	defer cleanup()

	// Enconde
	encodeExec, err := authtest.TxEncodeExec(val1.ClientCtx, savedTxFile.Name())
	s.Require().NoError(err)

	trimmedBase64 := strings.Trim(string(encodeExec), "\"\n")
	// Check that the transaction decodes as expected

	decodedTx, err := authtest.TxDecodeExec(val1.ClientCtx, trimmedBase64)
	s.Require().NoError(err)

	theTx := cli.UnmarshalStdTx(s.T(), val1.ClientCtx.JSONMarshaler, string(decodedTx))
	s.Require().Equal("deadbeef", theTx.Memo)
}

func TestCLIMultisignSortSignatures(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server with minimum fees
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	fooBarBazAddr := f.KeyAddress(cli.KeyFooBarBaz)
	barAddr := f.KeyAddress(cli.KeyBar)

	// Send some tokens from one account to the other
	success, _, _ := bankcli.TxSend(f, cli.KeyFoo, fooBarBazAddr, sdk.NewInt64Coin(cli.Denom, 10), "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, int64(10), bankcli.QueryBalances(f, fooBarBazAddr).AmountOf(cli.Denom).Int64())

	// Test generate sendTx with multisig
	success, stdout, _ := bankcli.TxSend(f, fooBarBazAddr.String(), barAddr, sdk.NewInt64Coin(cli.Denom, 5), "--generate-only")
	require.True(t, success)

	// Write the output to disk
	unsignedTxFile, cleanup := testutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with foo's key
	success, stdout, _ = authtest.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile, cleanup := testutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with baz's key
	success, stdout, _ = authtest.TxSign(f, cli.KeyBaz, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	bazSignatureFile, cleanup := testutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Multisign, keys in different order
	success, stdout, _ = authtest.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		bazSignatureFile.Name(), fooSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile, cleanup := testutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Validate the multisignature
	success, _, _ = authtest.TxValidateSignatures(f, signedTxFile.Name())
	require.True(t, success)

	// Broadcast the transaction
	success, _, _ = authtest.TxBroadcast(f, signedTxFile.Name())
	require.True(t, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIMultisign(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server with minimum fees
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	fooBarBazAddr := f.KeyAddress(cli.KeyFooBarBaz)
	bazAddr := f.KeyAddress(cli.KeyBaz)

	// Send some tokens from one account to the other
	success, _, _ := bankcli.TxSend(f, cli.KeyFoo, fooBarBazAddr, sdk.NewInt64Coin(cli.Denom, 10), "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, int64(10), bankcli.QueryBalances(f, fooBarBazAddr).AmountOf(cli.Denom).Int64())

	// Test generate sendTx with multisig
	success, stdout, stderr := bankcli.TxSend(f, fooBarBazAddr.String(), bazAddr, sdk.NewInt64Coin(cli.Denom, 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	// Write the output to disk
	unsignedTxFile, cleanup := testutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with foo's key
	success, stdout, _ = authtest.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile, cleanup := testutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with bar's key
	success, stdout, _ = authtest.TxSign(f, cli.KeyBar, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.True(t, success)

	// Write the output to disk
	barSignatureFile, cleanup := testutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Multisign

	// Does not work in offline mode
	success, stdout, _ = authtest.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		fooSignatureFile.Name(), barSignatureFile.Name()}, "--offline")
	require.Contains(t, "couldn't verify signature", stdout)
	require.False(t, success)

	// Success multisign
	success, stdout, _ = authtest.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		fooSignatureFile.Name(), barSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile, cleanup := testutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Validate the multisignature
	success, _, _ = authtest.TxValidateSignatures(f, signedTxFile.Name())
	require.True(t, success)

	// Broadcast the transaction
	success, _, _ = authtest.TxBroadcast(f, signedTxFile.Name())
	require.True(t, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
