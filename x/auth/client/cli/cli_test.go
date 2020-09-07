package cli_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codec2 "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtest "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

	kb := s.network.Validators[0].ClientCtx.Keyring
	_, _, err := kb.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	account1, _, err := kb.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	account2, _, err := kb.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	multi := multisig.NewPubKeyMultisigThreshold(2, []tmcrypto.PubKey{account1.GetPubKey(), account2.GetPubKey()})
	_, err = kb.SaveMultisig("multi", multi)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
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

	// write  unsigned tx to file
	unsignedTx, cleanup := testutil.WriteToNewTempFile(s.T(), res.String())
	defer cleanup()

	res, err = authtest.TxSignExec(val.ClientCtx, val.Address, unsignedTx.Name())
	s.Require().NoError(err)
	signedTx, err := val.ClientCtx.TxConfig.TxJSONDecoder()(res.Bytes())
	s.Require().NoError(err)

	signedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), res.String())
	defer cleanup()
	txBuilder, err := val.ClientCtx.TxConfig.WrapTxBuilder(signedTx)
	res, err = authtest.TxValidateSignaturesExec(val.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	txBuilder.SetMemo("MODIFIED TX")
	bz, err := val.ClientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
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
	filename, cleanup1 := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 3))
	defer cleanup1()

	// sign-batch file - offline is set but account-number and sequence are not
	val.ClientCtx.HomeDir = strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)
	res, err := authtest.TxSignBatchExec(val.ClientCtx, val.Address, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// sign-batch file
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// sign-batch file
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// Sign batch malformed tx file.
	malformedFile, cleanup2 := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf("%smalformed", generatedStd))
	defer cleanup2()

	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID))
	s.Require().Error(err)

	// Sign batch malformed tx file signature only.
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestCLITxQueryCmd() {
	val := s.network.Validators[0]
	var txHash string

	s.Run("bank send tx", func() {
		clientCtx := val.ClientCtx

		bz, err := bankcli.MsgSendExec(clientCtx, val.Address, val.Address, sdk.NewCoins(
			sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
			sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
		), []string{
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		}...)

		var txRes sdk.TxResponse
		s.Require().NoError(err)
		s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(bz.Bytes(), &txRes), bz.String())

		txHash = txRes.TxHash
		s.Require().Equal(uint32(0), txRes.Code)
	})

	s.network.WaitForNextBlock()

	s.Run("test QueryTxCmd", func() {
		cmd := authcli.QueryTxCmd()
		args := []string{
			txHash,
		}

		out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
		s.Require().NoError(err)

		var tx sdk.TxResponse
		s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &tx))
		fmt.Println(tx)
	})
}

func (s *IntegrationTestSuite) TestCLISendGenerateSignAndBroadcast() {
	val1 := s.network.Validators[0]

	account, err := val1.ClientCtx.Keyring.Key("newAccount")
	s.Require().NoError(err)

	sendTokens := sdk.NewCoin(s.cfg.BondDenom, sdk.TokensFromConsensusPower(10))

	normalGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		account.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	txCfg := val1.ClientCtx.TxConfig

	normalGeneratedStdTx, err := txCfg.TxJSONDecoder()(normalGeneratedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err := txCfg.WrapTxBuilder(normalGeneratedStdTx)
	s.Require().NoError(err)
	s.Require().Equal(txBuilder.GetTx().GetGas(), uint64(flags.DefaultGasLimit))
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err := txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(0, len(sigs))

	// Test generate sendTx with --gas=$amount
	limitedGasGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		account.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", 100),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	limitedGasStdTx, err := txCfg.TxJSONDecoder()(limitedGasGeneratedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err = txCfg.WrapTxBuilder(limitedGasStdTx)
	s.Require().NoError(err)
	s.Require().Equal(txBuilder.GetTx().GetGas(), uint64(100))
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(0, len(sigs))

	resp, err := bankcli.QueryBalancesExec(val1.ClientCtx, val1.Address)
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
	startTokens := balRes.Balances.AmountOf(s.cfg.BondDenom)

	// Test generate sendTx, estimate gas
	finalGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		account.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	finalStdTx, err := txCfg.TxJSONDecoder()(finalGeneratedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err = txCfg.WrapTxBuilder(finalStdTx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(flags.DefaultGasLimit), txBuilder.GetTx().GetGas())
	s.Require().Equal(len(finalStdTx.GetMsgs()), 1)

	// Write the output to disk
	unsignedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), finalGeneratedTx.String())
	defer cleanup()

	// Test validate-signatures
	res, err := authtest.TxValidateSignaturesExec(val1.ClientCtx, unsignedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
	s.Require().True(strings.Contains(res.String(), fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n\n", val1.Address.String())))

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
	signedFinalTx, err := txCfg.TxJSONDecoder()(signedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err = val1.ClientCtx.TxConfig.WrapTxBuilder(signedFinalTx)
	s.Require().NoError(err)
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(1, len(sigs))
	s.Require().Equal(val1.Address.String(), txBuilder.GetTx().GetSigners()[0].String())

	// Write the output to disk
	signedTxFile, cleanup2 := testutil.WriteToNewTempFile(s.T(), signedTx.String())
	defer cleanup2()

	// Validate Signature
	res, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)
	s.Require().True(strings.Contains(res.String(), "[OK]"))

	// Ensure foo has right amount of funds
	resp, err = bankcli.QueryBalancesExec(val1.ClientCtx, val1.Address)
	s.Require().NoError(err)

	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
	s.Require().Equal(startTokens, balRes.Balances.AmountOf(s.cfg.BondDenom))

	// Test broadcast

	// Does not work in offline mode
	res, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name(), "--offline")
	s.Require().EqualError(err, "cannot broadcast tx during offline mode")

	s.Require().NoError(s.network.WaitForNextBlock())

	// Broadcast correct transaction.
	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	res, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())

	// Ensure destiny account state
	resp, err = bankcli.QueryBalancesExec(val1.ClientCtx, account.GetAddress())
	s.Require().NoError(err)

	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
	s.Require().Equal(sendTokens.Amount, balRes.Balances.AmountOf(s.cfg.BondDenom))

	// Ensure origin account state
	resp, err = bankcli.QueryBalancesExec(val1.ClientCtx, val1.Address)
	s.Require().NoError(err)

	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestCLIMultisignInsufficientCosigners() {
	val1 := s.network.Validators[0]

	codec := codec2.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(codec)
	banktypes.RegisterLegacyAminoCodec(codec)
	val1.ClientCtx.LegacyAmino = codec

	// Generate 2 accounts and a multisig.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	multisigInfo, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	_, err = bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		multisigInfo.GetAddress(),
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 10),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())

	// Generate multisig transaction.
	multiGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		multisigInfo.GetAddress(),
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer cleanup()

	// Multisign, sign with one signature
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File, cleanup2 := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer cleanup2()

	multiSigWith1Signature, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name())
	s.Require().NoError(err)

	// Save tx to file
	multiSigWith1SignatureFile, cleanup3 := testutil.WriteToNewTempFile(s.T(), multiSigWith1Signature.String())
	defer cleanup3()

	exec, err := authtest.TxValidateSignaturesExec(val1.ClientCtx, multiSigWith1SignatureFile.Name())
	s.Require().Error(err)

	fmt.Printf("%s", exec)
}

func (s *IntegrationTestSuite) TestCLIEncode() {
	val1 := s.network.Validators[0]

	sendTokens := sdk.NewCoin(s.cfg.BondDenom, sdk.TokensFromConsensusPower(10))

	normalGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		val1.Address,
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly), "--memo", "deadbeef",
	)
	s.Require().NoError(err)

	// Save tx to file
	savedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
	defer cleanup()

	// Encode
	encodeExec, err := authtest.TxEncodeExec(val1.ClientCtx, savedTxFile.Name())
	s.Require().NoError(err)

	trimmedBase64 := strings.Trim(encodeExec.String(), "\"\n")
	// Check that the transaction decodes as expected

	decodedTx, err := authtest.TxDecodeExec(val1.ClientCtx, trimmedBase64)
	s.Require().NoError(err)

	txCfg := val1.ClientCtx.TxConfig
	theTx, err := txCfg.TxJSONDecoder()(decodedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err := val1.ClientCtx.TxConfig.WrapTxBuilder(theTx)
	s.Require().NoError(err)
	s.Require().Equal("deadbeef", txBuilder.GetTx().GetMemo())
}

func (s *IntegrationTestSuite) TestCLIMultisignSortSignatures() {
	val1 := s.network.Validators[0]

	codec := codec2.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(codec)
	banktypes.RegisterLegacyAminoCodec(codec)
	val1.ClientCtx.LegacyAmino = codec

	// Generate 2 accounts and a multisig.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := val1.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigInfo, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	resp, err := bankcli.QueryBalancesExec(val1.ClientCtx, multisigInfo.GetAddress())
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
	intialCoins := balRes.Balances

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		multisigInfo.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())

	resp, err = bankcli.QueryBalancesExec(val1.ClientCtx, multisigInfo.GetAddress())
	s.Require().NoError(err)

	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
	diff, _ := balRes.Balances.SafeSub(intialCoins)
	s.Require().Equal(sendTokens.Amount, diff.AmountOf(s.cfg.BondDenom))

	// Generate multisig transaction.
	multiGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		multisigInfo.GetAddress(),
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer cleanup()

	// Sign with account1
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File, cleanup2 := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer cleanup2()

	// Sign with account1
	account2Signature, err := authtest.TxSignExec(val1.ClientCtx, account2.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign2File, cleanup3 := testutil.WriteToNewTempFile(s.T(), account2Signature.String())
	defer cleanup3()

	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile, cleanup4 := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())
	defer cleanup4()

	_, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	_, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TestCLIMultisign() {
	val1 := s.network.Validators[0]

	codec := codec2.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(codec)
	banktypes.RegisterLegacyAminoCodec(codec)
	val1.ClientCtx.LegacyAmino = codec

	// Generate 2 accounts and a multisig.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := val1.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigInfo, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		multisigInfo.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)

	s.Require().NoError(s.network.WaitForNextBlock())

	resp, err := bankcli.QueryBalancesExec(val1.ClientCtx, multisigInfo.GetAddress())
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
	s.Require().Equal(sendTokens.Amount, balRes.Balances.AmountOf(s.cfg.BondDenom))

	// Generate multisig transaction.
	multiGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		multisigInfo.GetAddress(),
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile, cleanup := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer cleanup()

	// Sign with account1
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File, cleanup2 := testutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer cleanup2()

	// Sign with account1
	account2Signature, err := authtest.TxSignExec(val1.ClientCtx, account2.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign2File, cleanup3 := testutil.WriteToNewTempFile(s.T(), account2Signature.String())
	defer cleanup3()

	// Does not work in offline mode.
	_, err = authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), "--offline", sign1File.Name(), sign2File.Name())
	s.Require().EqualError(err, "couldn't verify signature: unable to verify single signer signature")

	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile, cleanup4 := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())
	defer cleanup4()

	_, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	_, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TestGetAccountCmd() {
	val := s.network.Validators[0]
	_, _, addr1 := testdata.KeyTestPubAddr()

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"invalid address",
			[]string{addr1.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
		},
		{
			"valid address",
			[]string{val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := authcli.GetAccountCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var any types.Any
				s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &any))
				var acc authtypes.AccountI
				s.Require().NoError(val.ClientCtx.InterfaceRegistry.UnpackAny(&any, &acc))
				s.Require().Equal(val.Address, acc.GetAddress())
			}
		})
	}
}

func TestGetBroadcastCommand_OfflineFlag(t *testing.T) {
	clientCtx := client.Context{}.WithOffline(true)
	clientCtx = clientCtx.WithTxConfig(simapp.MakeEncodingConfig().TxConfig)

	cmd := authcli.GetBroadcastCommand()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=true", flags.FlagOffline), ""})

	require.EqualError(t, cmd.Execute(), "cannot broadcast tx during offline mode")
}

func TestGetBroadcastCommand_WithoutOfflineFlag(t *testing.T) {
	clientCtx := client.Context{}
	txCfg := simapp.MakeEncodingConfig().TxConfig
	clientCtx = clientCtx.WithTxConfig(txCfg)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd := authcli.GetBroadcastCommand()
	_, out := testutil.ApplyMockIO(cmd)

	testDir, cleanFunc := testutil.NewTestCaseDir(t)
	t.Cleanup(cleanFunc)

	// Create new file with tx
	builder := txCfg.NewTxBuilder()
	builder.SetGasLimit(200000)
	from, err := sdk.AccAddressFromBech32("cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw")
	require.NoError(t, err)
	to, err := sdk.AccAddressFromBech32("cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw")
	require.NoError(t, err)
	err = builder.SetMsgs(banktypes.NewMsgSend(from, to, sdk.Coins{sdk.NewInt64Coin("stake", 10000)}))
	require.NoError(t, err)
	txContents, err := txCfg.TxJSONEncoder()(builder.GetTx())
	txFileName := filepath.Join(testDir, "tx.json")
	err = ioutil.WriteFile(txFileName, txContents, 0644)
	require.NoError(t, err)

	cmd.SetArgs([]string{txFileName})
	require.Error(t, cmd.ExecuteContext(ctx))
	require.Contains(t, out.String(), "unsupported return type ; supported types: sync, async, block")
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
