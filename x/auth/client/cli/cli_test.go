// +build norace

package cli_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
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
	cfg.NumValidators = 2

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	kb := s.network.Validators[0].ClientCtx.Keyring
	_, _, err := kb.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	account1, _, err := kb.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	account2, _, err := kb.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{account1.GetPubKey(), account2.GetPubKey()})
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
	sendTokens := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)))

	res, err := s.createBankMsg(val, val.Address, sendTokens,
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly))
	s.Require().NoError(err)

	// write  unsigned tx to file
	unsignedTx := testutil.WriteToNewTempFile(s.T(), res.String())
	res, err = authtest.TxSignExec(val.ClientCtx, val.Address, unsignedTx.Name())
	s.Require().NoError(err)
	signedTx, err := val.ClientCtx.TxConfig.TxJSONDecoder()(res.Bytes())
	s.Require().NoError(err)

	signedTxFile := testutil.WriteToNewTempFile(s.T(), res.String())
	txBuilder, err := val.ClientCtx.TxConfig.WrapTxBuilder(signedTx)
	res, err = authtest.TxValidateSignaturesExec(val.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	txBuilder.SetMemo("MODIFIED TX")
	bz, err := val.ClientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	modifiedTxFile := testutil.WriteToNewTempFile(s.T(), string(bz))

	res, err = authtest.TxValidateSignaturesExec(val.ClientCtx, modifiedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
}

func (s *IntegrationTestSuite) TestCLISignBatch() {
	val := s.network.Validators[0]
	var sendTokens = sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
	)

	generatedStd, err := s.createBankMsg(val, val.Address,
		sendTokens, fmt.Sprintf("--%s=true", flags.FlagGenerateOnly))
	s.Require().NoError(err)

	outputFile := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 3))
	val.ClientCtx.HomeDir = strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)

	// sign-batch file - offline is set but account-number and sequence are not
	res, err := authtest.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--offline")
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// sign-batch file
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// sign-batch file
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// Sign batch malformed tx file.
	malformedFile := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf("%smalformed", generatedStd))
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID))
	s.Require().Error(err)

	// Sign batch malformed tx file signature only.
	res, err = authtest.TxSignBatchExec(val.ClientCtx, val.Address, malformedFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--signature-only")
	s.Require().Error(err)
}

func (s *IntegrationTestSuite) TestCLISign_AminoJSON() {
	require := s.Require()
	val1 := s.network.Validators[0]
	txCfg := val1.ClientCtx.TxConfig
	var sendTokens = sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val1.Moniker), sdk.NewInt(10)),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
	)
	txBz, err := s.createBankMsg(val1, val1.Address,
		sendTokens, fmt.Sprintf("--%s=true", flags.FlagGenerateOnly))
	require.NoError(err)
	fileUnsigned := testutil.WriteToNewTempFile(s.T(), txBz.String())
	chainFlag := fmt.Sprintf("--%s=%s", flags.FlagChainID, val1.ClientCtx.ChainID)
	sigOnlyFlag := "--signature-only"
	signModeAminoFlag := "--sign-mode=amino-json"

	// SIC! validators have same key names and same adddresses as those registered in the keyring,
	//      BUT the keys are different!
	valInfo, err := val1.ClientCtx.Keyring.Key(val1.Moniker)
	require.NoError(err)

	// query account info
	queryResJSON, err := authtest.QueryAccountExec(val1.ClientCtx, val1.Address)
	require.NoError(err)
	var account authtypes.AccountI
	require.NoError(val1.ClientCtx.JSONMarshaler.UnmarshalInterfaceJSON(queryResJSON.Bytes(), &account))

	/****  test signature-only  ****/
	res, err := authtest.TxSignExec(val1.ClientCtx, val1.Address, fileUnsigned.Name(), chainFlag,
		sigOnlyFlag, signModeAminoFlag)
	require.NoError(err)
	checkSignatures(require, txCfg, res.Bytes(), valInfo.GetPubKey())
	sigs, err := txCfg.UnmarshalSignatureJSON(res.Bytes())
	require.NoError(err)
	require.Equal(1, len(sigs))
	require.Equal(account.GetSequence(), sigs[0].Sequence)

	/****  test full output  ****/
	res, err = authtest.TxSignExec(val1.ClientCtx, val1.Address, fileUnsigned.Name(), chainFlag, signModeAminoFlag)
	require.NoError(err)

	// txCfg.UnmarshalSignatureJSON can't unmarshal a fragment of the signature, so we create this structure.
	type txFragment struct {
		Signatures []json.RawMessage
	}
	var txOut txFragment
	err = json.Unmarshal(res.Bytes(), &txOut)
	require.NoError(err)
	require.Len(txOut.Signatures, 1)

	/****  test file output  ****/
	filenameSigned := filepath.Join(s.T().TempDir(), "test_sign_out.json")
	fileFlag := fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, filenameSigned)
	_, err = authtest.TxSignExec(val1.ClientCtx, val1.Address, fileUnsigned.Name(), chainFlag, fileFlag, signModeAminoFlag)
	require.NoError(err)
	fContent, err := ioutil.ReadFile(filenameSigned)
	require.NoError(err)
	require.Equal(res.String(), string(fContent))

	/****  try to append to the previously signed transaction  ****/
	res, err = authtest.TxSignExec(val1.ClientCtx, val1.Address, filenameSigned, chainFlag,
		sigOnlyFlag, signModeAminoFlag)
	require.NoError(err)
	checkSignatures(require, txCfg, res.Bytes(), valInfo.GetPubKey(), valInfo.GetPubKey())

	/****  try to overwrite the previously signed transaction  ****/

	// We can't sign with other address, because the bank send message supports only one signer for a simple
	// account. Changing the file is too much hacking, because TxDecoder returns sdk.Tx, which doesn't
	// provide functionality to check / manage `auth_info`.
	// Cases with different keys are are covered in unit tests of `tx.Sign`.
	res, err = authtest.TxSignExec(val1.ClientCtx, val1.Address, filenameSigned, chainFlag,
		sigOnlyFlag, "--overwrite", signModeAminoFlag)
	checkSignatures(require, txCfg, res.Bytes(), valInfo.GetPubKey())

	/****  test flagAmino  ****/
	res, err = authtest.TxSignExec(val1.ClientCtx, val1.Address, filenameSigned, chainFlag,
		"--amino=true", signModeAminoFlag)
	require.NoError(err)

	var txAmino authrest.BroadcastReq
	err = val1.ClientCtx.LegacyAmino.UnmarshalJSON(res.Bytes(), &txAmino)
	require.NoError(err)
	require.Len(txAmino.Tx.Signatures, 2)
	require.Equal(txAmino.Tx.Signatures[0].PubKey, valInfo.GetPubKey())
	require.Equal(txAmino.Tx.Signatures[1].PubKey, valInfo.GetPubKey())
}

func checkSignatures(require *require.Assertions, txCfg client.TxConfig, output []byte, pks ...cryptotypes.PubKey) {
	sigs, err := txCfg.UnmarshalSignatureJSON(output)
	require.NoError(err, string(output))
	require.Len(sigs, len(pks))
	for i := range pks {
		require.True(sigs[i].PubKey.Equals(pks[i]), "Pub key doesn't match. Got: %s, expected: %s, idx: %d", sigs[i].PubKey, pks[i], i)
		require.NotEmpty(sigs[i].Data)
	}
}

func (s *IntegrationTestSuite) TestCLIQueryTxCmdByHash() {
	val := s.network.Validators[0]

	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	// Send coins, try both with legacy Msg and with Msg service.
	// Legacy Msg.
	legacyMsgOut, err := bankcli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		account2.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)
	s.Require().NoError(err)
	var legacyMsgTxRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(legacyMsgOut.Bytes(), &legacyMsgTxRes))

	// Service Msg.
	out, err := bankcli.ServiceMsgSendExec(
		val.ClientCtx,
		val.Address,
		account2.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txRes))

	s.Require().NoError(s.network.WaitForNextBlock())

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		rawLogContains string
	}{
		{
			"not enough args",
			[]string{},
			true, "",
		},
		{
			"with invalid hash",
			[]string{"somethinginvalid", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true, "",
		},
		{
			"with valid and not existing hash",
			[]string{"C7E7D3A86A17AB3A321172239F3B61357937AF0F25D9FA4D2F4DCCAD9B0D7747", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true, "",
		},
		{
			"happy case (legacy Msg)",
			[]string{legacyMsgTxRes.TxHash, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
		},
		{
			"happy case (service Msg)",
			[]string{txRes.TxHash, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"/cosmos.bank.v1beta1.Msg/Send",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := authcli.QueryTxCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var result sdk.TxResponse
				s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &result))
				s.Require().NotNil(result.Height)
				s.Require().Contains(result.RawLog, tc.rawLogContains)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestCLIQueryTxCmdByEvents() {
	val := s.network.Validators[0]

	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	// Send coins.
	out, err := s.createBankMsg(
		val, account2.GetAddress(),
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().NoError(s.network.WaitForNextBlock())

	// Query the tx by hash to get the inner tx.
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, authcli.QueryTxCmd(), []string{txRes.TxHash, fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txRes))
	protoTx := txRes.GetTx().(*tx.Tx)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrStr string
	}{
		{
			"invalid --type",
			[]string{
				fmt.Sprintf("--type=%s", "foo"),
				"bar",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, "unknown --type value foo",
		},
		{
			"--type=acc_seq with no addr+seq",
			[]string{
				"--type=acc_seq",
				"",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, "`acc_seq` type takes an argument '<addr>/<seq>'",
		},
		{
			"non-existing addr+seq combo",
			[]string{
				"--type=acc_seq",
				"foobar",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, "found no txs matching given address and sequence combination",
		},
		{
			"addr+seq happy case",
			[]string{
				"--type=acc_seq",
				fmt.Sprintf("%s/%d", val.Address, protoTx.AuthInfo.SignerInfos[0].Sequence),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false, "",
		},
		{
			"--type=signature with no signature",
			[]string{
				"--type=signature",
				"",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, "argument should be comma-separated signatures",
		},
		{
			"non-existing signatures",
			[]string{
				"--type=signature",
				"foo",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true, "found no txs matching given signatures",
		},
		{
			"with --signatures happy case",
			[]string{
				"--type=signature",
				base64.StdEncoding.EncodeToString(protoTx.Signatures[0]),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false, "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := authcli.QueryTxCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrStr)
			} else {
				var result sdk.TxResponse
				s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &result))
				s.Require().NotNil(result.Height)
			}
		})
	}
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
	unsignedTxFile := testutil.WriteToNewTempFile(s.T(), finalGeneratedTx.String())

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
	signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx.String())

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
	val1 := *s.network.Validators[0]

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
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())

	// Multisign, sign with one signature
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())

	multiSigWith1Signature, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name())
	s.Require().NoError(err)

	// Save tx to file
	multiSigWith1SignatureFile := testutil.WriteToNewTempFile(s.T(), multiSigWith1Signature.String())

	_, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, multiSigWith1SignatureFile.Name())
	s.Require().Error(err)
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
	savedTxFile := testutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())

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
	val1 := *s.network.Validators[0]

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
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())

	// Sign with account1
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())

	// Sign with account1
	account2Signature, err := authtest.TxSignExec(val1.ClientCtx, account2.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())

	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())

	_, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	_, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TestCLIMultisign() {
	val1 := *s.network.Validators[0]

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
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())

	// Sign with account1
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, account1.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())

	// Sign with account2
	account2Signature, err := authtest.TxSignExec(val1.ClientCtx, account2.GetAddress(), multiGeneratedTxFile.Name(), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)

	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())

	// Does not work in offline mode.
	_, err = authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), "--offline", sign1File.Name(), sign2File.Name())
	s.Require().EqualError(err, "couldn't verify signature: unable to verify single signer signature")

	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())

	_, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	_, err = authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TestSignBatchMultisig() {
	val := s.network.Validators[0]

	// Fetch 2 accounts and a multisig.
	account1, err := val.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)
	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)
	multisigInfo, err := val.ClientCtx.Keyring.Key("multi")

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, err = bankcli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		multisigInfo.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	generatedStd, err := bankcli.MsgSendExec(
		val.ClientCtx,
		multisigInfo.GetAddress(),
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(1)),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Write the output to disk
	filename := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 1))
	val.ClientCtx.HomeDir = strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)

	// sign-batch file
	res, err := authtest.TxSignBatchExec(val.ClientCtx, account1.GetAddress(), filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)
	s.Require().Equal(1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file
	file1 := testutil.WriteToNewTempFile(s.T(), res.String())

	// sign-batch file with account2
	res, err = authtest.TxSignBatchExec(val.ClientCtx, account2.GetAddress(), filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--multisig", multisigInfo.GetAddress().String())
	s.Require().NoError(err)
	s.Require().Equal(1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// write sigs to file2
	file2 := testutil.WriteToNewTempFile(s.T(), res.String())
	res, err = authtest.TxMultiSignExec(val.ClientCtx, multisigInfo.GetName(), filename.Name(), file1.Name(), file2.Name())
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestMultisignBatch() {
	val := s.network.Validators[0]

	// Fetch 2 accounts and a multisig.
	account1, err := val.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)
	account2, err := val.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)
	multisigInfo, err := val.ClientCtx.Keyring.Key("multi")

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 1000)
	_, err = bankcli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		multisigInfo.GetAddress(),
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	generatedStd, err := bankcli.MsgSendExec(
		val.ClientCtx,
		multisigInfo.GetAddress(),
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(1)),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)

	// Write the output to disk
	filename := testutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 3))
	val.ClientCtx.HomeDir = strings.Replace(val.ClientCtx.HomeDir, "simd", "simcli", 1)

	queryResJSON, err := authtest.QueryAccountExec(val.ClientCtx, multisigInfo.GetAddress())
	s.Require().NoError(err)
	var account authtypes.AccountI
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalInterfaceJSON(queryResJSON.Bytes(), &account))

	// sign-batch file
	res, err := authtest.TxSignBatchExec(val.ClientCtx, account1.GetAddress(), filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--multisig", multisigInfo.GetAddress().String(), fmt.Sprintf("--%s", flags.FlagOffline), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, fmt.Sprint(account.GetAccountNumber())), fmt.Sprintf("--%s=%s", flags.FlagSequence, fmt.Sprint(account.GetSequence())))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file
	file1 := testutil.WriteToNewTempFile(s.T(), res.String())

	// sign-batch file with account2
	res, err = authtest.TxSignBatchExec(val.ClientCtx, account2.GetAddress(), filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, val.ClientCtx.ChainID), "--multisig", multisigInfo.GetAddress().String(), fmt.Sprintf("--%s", flags.FlagOffline), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, fmt.Sprint(account.GetAccountNumber())), fmt.Sprintf("--%s=%s", flags.FlagSequence, fmt.Sprint(account.GetSequence())))
	s.Require().NoError(err)
	s.Require().Equal(3, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))

	// multisign the file
	file2 := testutil.WriteToNewTempFile(s.T(), res.String())
	res, err = authtest.TxMultiSignBatchExec(val.ClientCtx, filename.Name(), multisigInfo.GetName(), file1.Name(), file2.Name())
	s.Require().NoError(err)
	signedTxs := strings.Split(strings.Trim(res.String(), "\n"), "\n")

	// Broadcast transactions.
	for _, signedTx := range signedTxs {
		signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx)
		val.ClientCtx.BroadcastMode = flags.BroadcastBlock
		res, err = authtest.TxBroadcastExec(val.ClientCtx, signedTxFile.Name())
		s.T().Log(res)
		s.Require().NoError(err)
		s.Require().NoError(s.network.WaitForNextBlock())
	}
}

func (s *IntegrationTestSuite) TestGetAccountCmd() {
	val := s.network.Validators[0]
	_, _, addr1 := testdata.KeyTestPubAddr()

	testCases := []struct {
		name      string
		address   sdk.AccAddress
		expectErr bool
	}{
		{
			"invalid address",
			addr1,
			true,
		},
		{
			"valid address",
			val.Address,
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			out, err := authtest.QueryAccountExec(clientCtx, tc.address)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var acc authtypes.AccountI
				s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalInterfaceJSON(out.Bytes(), &acc))
				s.Require().Equal(val.Address, acc.GetAddress())
			}
		})
	}
}

func TestGetBroadcastCommand_OfflineFlag(t *testing.T) {
	clientCtx := client.Context{}.WithOffline(true)
	clientCtx = clientCtx.WithTxConfig(simapp.MakeTestEncodingConfig().TxConfig)

	cmd := authcli.GetBroadcastCommand()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=true", flags.FlagOffline), ""})

	require.EqualError(t, cmd.Execute(), "cannot broadcast tx during offline mode")
}

func TestGetBroadcastCommand_WithoutOfflineFlag(t *testing.T) {
	clientCtx := client.Context{}
	txCfg := simapp.MakeTestEncodingConfig().TxConfig
	clientCtx = clientCtx.WithTxConfig(txCfg)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd := authcli.GetBroadcastCommand()
	_, out := testutil.ApplyMockIO(cmd)

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
	require.NoError(t, err)
	txFile := testutil.WriteToNewTempFile(t, string(txContents))

	cmd.SetArgs([]string{txFile.Name()})
	err = cmd.ExecuteContext(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connect: connection refused")
	require.Contains(t, out.String(), "connect: connection refused")
}

func (s *IntegrationTestSuite) TestQueryParamsCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"happy case",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
		},
		{
			"with specific height",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := authcli.QueryParamsCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var authParams authtypes.Params
				s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &authParams))
				s.Require().NotNil(authParams.MaxMemoCharacters)
			}
		})
	}
}

// TestTxWithoutPublicKey makes sure sending a proto tx message without the
// public key doesn't cause any error in the RPC layer (broadcast).
// See https://github.com/cosmos/cosmos-sdk/issues/7585 for more details.
func (s *IntegrationTestSuite) TestTxWithoutPublicKey() {
	val1 := s.network.Validators[0]
	txCfg := val1.ClientCtx.TxConfig

	// Create a txBuilder with an unsigned tx.
	txBuilder := txCfg.NewTxBuilder()
	msg := banktypes.NewMsgSend(val1.Address, val1.Address, sdk.NewCoins(
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
	))
	err := txBuilder.SetMsgs(msg)
	s.Require().NoError(err)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(150))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	// Set empty signature to set signer infos.
	sigV2 := signing.SignatureV2{
		PubKey: val1.PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  txCfg.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
	}
	err = txBuilder.SetSignatures(sigV2)
	s.Require().NoError(err)

	// Create a file with the unsigned tx.
	txJSON, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	unsignedTxFile := testutil.WriteToNewTempFile(s.T(), string(txJSON))

	// Sign the file with the unsignedTx.
	signedTx, err := authtest.TxSignExec(val1.ClientCtx, val1.Address, unsignedTxFile.Name())
	s.Require().NoError(err)

	// Remove the signerInfo's `public_key` field manually from the signedTx.
	// Note: this method is only used for test purposes! In general, one should
	// use txBuilder and TxEncoder/TxDecoder to manipulate txs.
	var tx tx.Tx
	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(signedTx.Bytes(), &tx)
	s.Require().NoError(err)
	tx.AuthInfo.SignerInfos[0].PublicKey = nil
	// Re-encode the tx again, to another file.
	txJSON, err = val1.ClientCtx.JSONMarshaler.MarshalJSON(&tx)
	s.Require().NoError(err)
	signedTxFile := testutil.WriteToNewTempFile(s.T(), string(txJSON))
	s.Require().True(strings.Contains(string(txJSON), "\"public_key\":null"))

	// Broadcast tx, test that it shouldn't panic.
	val1.ClientCtx.BroadcastMode = flags.BroadcastSync
	out, err := authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)
	var res sdk.TxResponse
	s.Require().NoError(val1.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
	s.Require().NotEqual(0, res.Code)
}

func (s *IntegrationTestSuite) TestSignWithMultiSigners_AminoJSON() {
	// test case:
	// Create a transaction with 2 messages which has to be signed with 2 different keys
	// Sign and append the signatures using the CLI with Amino signing mode.
	// Finally send the transaction to the blockchain. It should work.

	require := s.Require()
	val0, val1 := s.network.Validators[0], s.network.Validators[1]
	val0Coin := sdk.NewCoin(fmt.Sprintf("%stoken", val0.Moniker), sdk.NewInt(10))
	val1Coin := sdk.NewCoin(fmt.Sprintf("%stoken", val1.Moniker), sdk.NewInt(10))
	_, _, addr1 := testdata.KeyTestPubAddr()

	// Creating a tx with 2 msgs from 2 signers: val0 and val1.
	// The validators need to sign with SIGN_MODE_LEGACY_AMINO_JSON,
	// because DIRECT doesn't support multi signers via the CLI.
	// Since we we amino, we don't need to pre-populate signer_infos.
	txBuilder := val0.ClientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(
		banktypes.NewMsgSend(val0.Address, addr1, sdk.NewCoins(val0Coin)),
		banktypes.NewMsgSend(val1.Address, addr1, sdk.NewCoins(val1Coin)),
	)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	require.Equal([]sdk.AccAddress{val0.Address, val1.Address}, txBuilder.GetTx().GetSigners())

	// Write the unsigned tx into a file.
	txJSON, err := val0.ClientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(err)
	unsignedTxFile := testutil.WriteToNewTempFile(s.T(), string(txJSON))

	// Let val0 sign first the file with the unsignedTx.
	signedByVal0, err := authtest.TxSignExec(val0.ClientCtx, val0.Address, unsignedTxFile.Name(), "--overwrite", "--sign-mode=amino-json")
	require.NoError(err)
	signedByVal0File := testutil.WriteToNewTempFile(s.T(), signedByVal0.String())

	// Then let val1 sign the file with signedByVal0.
	val1AccNum, val1Seq, err := val0.ClientCtx.AccountRetriever.GetAccountNumberSequence(val0.ClientCtx, val1.Address)
	require.NoError(err)
	signedTx, err := authtest.TxSignExec(
		val1.ClientCtx, val1.Address, signedByVal0File.Name(),
		"--offline", fmt.Sprintf("--account-number=%d", val1AccNum), fmt.Sprintf("--sequence=%d", val1Seq), "--sign-mode=amino-json",
	)
	require.NoError(err)
	signedTxFile := testutil.WriteToNewTempFile(s.T(), signedTx.String())

	// Now let's try to send this tx.
	res, err := authtest.TxBroadcastExec(val0.ClientCtx, signedTxFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock))
	require.NoError(err)
	var txRes sdk.TxResponse
	require.NoError(val0.ClientCtx.JSONMarshaler.UnmarshalJSON(res.Bytes(), &txRes))
	require.Equal(uint32(0), txRes.Code)

	// Make sure the addr1's balance got funded.
	queryResJSON, err := bankcli.QueryBalancesExec(val0.ClientCtx, addr1)
	require.NoError(err)
	var queryRes banktypes.QueryAllBalancesResponse
	err = val0.ClientCtx.JSONMarshaler.UnmarshalJSON(queryResJSON.Bytes(), &queryRes)
	require.NoError(err)
	require.Equal(sdk.NewCoins(val0Coin, val1Coin), queryRes.Balances)
}

func (s *IntegrationTestSuite) createBankMsg(val *network.Validator, toAddr sdk.AccAddress, amount sdk.Coins, extraFlags ...string) (testutil.BufferWriter, error) {
	flags := []string{fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees,
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	flags = append(flags, extraFlags...)
	return bankcli.MsgSendExec(val.ClientCtx, val.Address, toAddr, amount, flags...)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
