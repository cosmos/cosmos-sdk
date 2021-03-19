// +build norace

package rest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authtest "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	ibccli "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/client/cli"
	ibcsolomachinecli "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	stdTx    legacytx.StdTx
	stdTxRes sdk.TxResponse
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

	// Broadcast a StdTx used for tests.
	s.stdTx = s.createTestStdTx(s.network.Validators[0], 0, 1)
	res, err := s.broadcastReq(s.stdTx, "block")
	s.Require().NoError(err)

	// NOTE: this uses amino explicitly, don't migrate it!
	s.Require().NoError(s.cfg.LegacyAmino.UnmarshalJSON(res, &s.stdTxRes))
	s.Require().Equal(uint32(0), s.stdTxRes.Code)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func mkStdTx() legacytx.StdTx {
	// NOTE: this uses StdTx explicitly, don't migrate it!
	return legacytx.StdTx{
		Msgs: []sdk.Msg{&types.MsgSend{}},
		Fee: legacytx.StdFee{
			Amount: sdk.Coins{sdk.NewInt64Coin("foo", 10)},
			Gas:    10000,
		},
		Memo: "FOOBAR",
	}
}

// Create an IBC tx that's encoded as amino-JSON. Since we can't amino-marshal
// a tx with "cosmos-sdk/MsgTransfer" using the SDK, we just hardcode the tx
// here. But external clients might, see https://github.com/cosmos/cosmos-sdk/issues/8022.
func mkIBCStdTx() []byte {
	ibcTx := `{
		"account_number": "68",
		"chain_id": "stargate-4",
		"fee": {
		  "amount": [
			{
			  "amount": "3500",
			  "denom": "umuon"
			}
		  ],
		  "gas": "350000"
		},
		"memo": "",
		"msg": [
		  {
			"type": "cosmos-sdk/MsgTransfer",
			"value": {
			  "receiver": "cosmos1q9wtnlwdjrhwtcjmt2uq77jrgx7z3usrq2yz7z",
			  "sender": "cosmos1q9wtnlwdjrhwtcjmt2uq77jrgx7z3usrq2yz7z",
			  "source_channel": "THEslipperCHANNEL",
			  "source_port": "transfer",
			  "token": {
				"amount": "1000000",
				"denom": "umuon"
			  }
			}
		  }
		],
		"sequence": "24"
	  }`
	req := fmt.Sprintf(`{"tx":%s,"mode":"async"}`, ibcTx)

	return []byte(req)
}

func (s *IntegrationTestSuite) TestEncodeDecode() {
	var require = s.Require()
	val := s.network.Validators[0]
	stdTx := mkStdTx()

	// NOTE: this uses amino explicitly, don't migrate it!
	cdc := val.ClientCtx.LegacyAmino

	bz, err := cdc.MarshalJSON(stdTx)
	require.NoError(err)

	res, err := rest.PostRequest(fmt.Sprintf("%s/txs/encode", val.APIAddress), "application/json", bz)
	require.NoError(err)

	var encodeResp authrest.EncodeResp
	err = cdc.UnmarshalJSON(res, &encodeResp)
	require.NoError(err)

	bz, err = cdc.MarshalJSON(authrest.DecodeReq{Tx: encodeResp.Tx})
	require.NoError(err)

	res, err = rest.PostRequest(fmt.Sprintf("%s/txs/decode", val.APIAddress), "application/json", bz)
	require.NoError(err)

	var respWithHeight rest.ResponseWithHeight
	err = cdc.UnmarshalJSON(res, &respWithHeight)
	require.NoError(err)
	var decodeResp authrest.DecodeResp
	err = cdc.UnmarshalJSON(respWithHeight.Result, &decodeResp)
	require.NoError(err)
	require.Equal(stdTx, legacytx.StdTx(decodeResp))
}

func (s *IntegrationTestSuite) TestEncodeIBCTx() {
	val := s.network.Validators[0]

	req := mkIBCStdTx()
	res, err := rest.PostRequest(fmt.Sprintf("%s/txs/encode", val.APIAddress), "application/json", []byte(req))
	s.Require().NoError(err)

	s.Require().Contains(string(res), authrest.ErrEncodeDecode.Error())
}

func (s *IntegrationTestSuite) TestBroadcastTxRequest() {
	stdTx := mkStdTx()

	// we just test with async mode because this tx will fail - all we care about is that it got encoded and broadcast correctly
	res, err := s.broadcastReq(stdTx, "async")
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	// NOTE: this uses amino explicitly, don't migrate it!
	s.Require().NoError(s.cfg.LegacyAmino.UnmarshalJSON(res, &txRes))
	// we just check for a non-empty TxHash here, the actual hash will depend on the underlying tx configuration
	s.Require().NotEmpty(txRes.TxHash)
}

func (s *IntegrationTestSuite) TestBroadcastIBCTxRequest() {
	val := s.network.Validators[0]

	req := mkIBCStdTx()
	res, err := rest.PostRequest(fmt.Sprintf("%s/txs", val.APIAddress), "application/json", []byte(req))
	s.Require().NoError(err)

	s.Require().NotContains(string(res), "this transaction cannot be broadcasted via legacy REST endpoints", string(res))
}

// Helper function to test querying txs. We will use it to query StdTx and service `Msg`s.
func (s *IntegrationTestSuite) testQueryTx(txHeight int64, txHash, txRecipient string) {
	val0 := s.network.Validators[0]

	testCases := []struct {
		desc     string
		malleate func() *sdk.TxResponse
	}{
		{
			"Query by hash",
			func() *sdk.TxResponse {
				txJSON, err := rest.GetRequest(fmt.Sprintf("%s/txs/%s", val0.APIAddress, txHash))
				s.Require().NoError(err)

				var txResAmino sdk.TxResponse
				s.Require().NoError(val0.ClientCtx.LegacyAmino.UnmarshalJSON(txJSON, &txResAmino))
				return &txResAmino
			},
		},
		{
			"Query by height",
			func() *sdk.TxResponse {
				txJSON, err := rest.GetRequest(fmt.Sprintf("%s/txs?limit=10&page=1&tx.height=%d", val0.APIAddress, txHeight))
				s.Require().NoError(err)

				var searchtxResult sdk.SearchTxsResult
				s.Require().NoError(val0.ClientCtx.LegacyAmino.UnmarshalJSON(txJSON, &searchtxResult))
				s.Require().Len(searchtxResult.Txs, 1)
				return searchtxResult.Txs[0]
			},
		},
		{
			"Query by event (transfer.recipient)",
			func() *sdk.TxResponse {
				txJSON, err := rest.GetRequest(fmt.Sprintf("%s/txs?transfer.recipient=%s", val0.APIAddress, txRecipient))
				s.Require().NoError(err)

				var searchtxResult sdk.SearchTxsResult
				s.Require().NoError(val0.ClientCtx.LegacyAmino.UnmarshalJSON(txJSON, &searchtxResult))
				s.Require().Len(searchtxResult.Txs, 1)
				return searchtxResult.Txs[0]
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			txResponse := tc.malleate()

			// Check that the height is correct.
			s.Require().Equal(txHeight, txResponse.Height)

			// Check that the events are correct.
			s.Require().Contains(
				txResponse.RawLog,
				fmt.Sprintf("{\"key\":\"recipient\",\"value\":\"%s\"}", txRecipient),
			)

			// Check that the Msg is correct.
			stdTx, ok := txResponse.Tx.GetCachedValue().(legacytx.StdTx)
			s.Require().True(ok)
			msgs := stdTx.GetMsgs()
			s.Require().Equal(len(msgs), 1)
			msg, ok := msgs[0].(*types.MsgSend)
			s.Require().True(ok)
			s.Require().Equal(txRecipient, msg.ToAddress)
		})
	}
}

func (s *IntegrationTestSuite) TestQueryTxWithStdTx() {
	val0 := s.network.Validators[0]

	// We broadcasted a StdTx in SetupSuite.
	// We just check for a non-empty TxHash here, the actual hash will depend on the underlying tx configuration
	s.Require().NotEmpty(s.stdTxRes.TxHash)

	s.testQueryTx(s.stdTxRes.Height, s.stdTxRes.TxHash, val0.Address.String())
}

func (s *IntegrationTestSuite) TestQueryTxWithServiceMessage() {
	val := s.network.Validators[0]

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)
	_, _, addr := testdata.KeyTestPubAddr()

	// Might need to wait a block to refresh sequences from previous setups.
	s.Require().NoError(s.network.WaitForNextBlock())

	out, err := bankcli.ServiceMsgSendExec(
		val.ClientCtx,
		val.Address,
		addr,
		sdk.NewCoins(sendTokens),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)

	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().Equal(uint32(0), txRes.Code)

	s.Require().NoError(s.network.WaitForNextBlock())

	s.testQueryTx(txRes.Height, txRes.TxHash, addr.String())
}

func (s *IntegrationTestSuite) TestMultipleSyncBroadcastTxRequests() {
	// First test transaction from validator should have sequence=1 (non-genesis tx)
	testCases := []struct {
		desc      string
		sequence  uint64
		shouldErr bool
	}{
		{
			"First tx (correct sequence)",
			1,
			false,
		},
		{
			"Second tx (correct sequence)",
			2,
			false,
		},
		{
			"Third tx (incorrect sequence)",
			9,
			true,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// broadcast test with sync mode, as we want to run CheckTx to verify account sequence is correct
			stdTx := s.createTestStdTx(s.network.Validators[1], 1, tc.sequence)
			res, err := s.broadcastReq(stdTx, "sync")
			s.Require().NoError(err)

			var txRes sdk.TxResponse
			// NOTE: this uses amino explicitly, don't migrate it!
			s.Require().NoError(s.cfg.LegacyAmino.UnmarshalJSON(res, &txRes))
			// we check for a exitCode=0, indicating a successful broadcast
			if tc.shouldErr {
				var sigVerifyFailureCode uint32 = 4
				s.Require().Equal(sigVerifyFailureCode, txRes.Code,
					"Testcase '%s': Expected signature verification failure {Code: %d} from TxResponse. "+
						"Found {Code: %d, RawLog: '%v'}",
					tc.desc, sigVerifyFailureCode, txRes.Code, txRes.RawLog,
				)
			} else {
				s.Require().Equal(uint32(0), txRes.Code,
					"Testcase '%s': TxResponse errored unexpectedly. Err: {Code: %d, RawLog: '%v'}",
					tc.desc, txRes.Code, txRes.RawLog,
				)
			}
		})
	}
}

func (s *IntegrationTestSuite) createTestStdTx(val *network.Validator, accNum, sequence uint64) legacytx.StdTx {
	txConfig := legacytx.StdTxConfig{Cdc: s.cfg.LegacyAmino}

	msg := &types.MsgSend{
		FromAddress: val.Address.String(),
		ToAddress:   val.Address.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin(fmt.Sprintf("%stoken", val.Moniker), 100)},
	}

	// prepare txBuilder with msg
	txBuilder := txConfig.NewTxBuilder()
	feeAmount := sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 10)}
	gasLimit := testdata.NewTestGasLimit()
	txBuilder.SetMsgs(msg)
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetMemo("foobar")

	// setup txFactory
	txFactory := tx.Factory{}.
		WithChainID(val.ClientCtx.ChainID).
		WithKeybase(val.ClientCtx.Keyring).
		WithTxConfig(txConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON).
		WithAccountNumber(accNum).
		WithSequence(sequence)

	// sign Tx (offline mode so we can manually set sequence number)
	err := authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, true, true)
	s.Require().NoError(err)

	stdTx := txBuilder.GetTx().(legacytx.StdTx)

	return stdTx
}

func (s *IntegrationTestSuite) broadcastReq(stdTx legacytx.StdTx, mode string) ([]byte, error) {
	val := s.network.Validators[0]

	// NOTE: this uses amino explicitly, don't migrate it!
	cdc := val.ClientCtx.LegacyAmino
	req := authrest.BroadcastReq{
		Tx:   stdTx,
		Mode: mode,
	}
	bz, err := cdc.MarshalJSON(req)
	s.Require().NoError(err)

	return rest.PostRequest(fmt.Sprintf("%s/txs", val.APIAddress), "application/json", bz)
}

// testQueryIBCTx is a helper function to test querying txs which:
// - show an error message on legacy REST endpoints
// - succeed using gRPC
// In practice, we call this function on IBC txs.
func (s *IntegrationTestSuite) testQueryIBCTx(txRes sdk.TxResponse, cmd *cobra.Command, args []string) {
	val := s.network.Validators[0]

	errMsg := "this transaction cannot be displayed via legacy REST endpoints, because it does not support" +
		" Amino serialization. Please either use CLI, gRPC, gRPC-gateway, or directly query the Tendermint RPC" +
		" endpoint to query this transaction. The new REST endpoint (via gRPC-gateway) is "

	// Test that legacy endpoint return the above error message on IBC txs.
	testCases := []struct {
		desc string
		url  string
	}{
		{
			"Query by hash",
			fmt.Sprintf("%s/txs/%s", val.APIAddress, txRes.TxHash),
		},
		{
			"Query by height",
			fmt.Sprintf("%s/txs?tx.height=%d", val.APIAddress, txRes.Height),
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			txJSON, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var errResp rest.ErrorResponse
			s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(txJSON, &errResp))

			s.Require().Contains(errResp.Error, errMsg)
		})
	}

	// try fetching the txn using gRPC req, it will fetch info since it has proto codec.
	grpcJSON, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", val.APIAddress, txRes.TxHash))
	s.Require().NoError(err)

	var getTxRes txtypes.GetTxResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(grpcJSON, &getTxRes))
	s.Require().Equal(getTxRes.Tx.Body.Memo, "foobar")

	// generate broadcast only txn.
	args = append(args, fmt.Sprintf("--%s=true", flags.FlagGenerateOnly))
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)

	txFile := testutil.WriteToNewTempFile(s.T(), string(out.Bytes()))
	txFileName := txFile.Name()

	// encode the generated txn.
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, authcli.GetEncodeCommand(), []string{txFileName})
	s.Require().NoError(err)

	bz, err := val.ClientCtx.LegacyAmino.MarshalJSON(authrest.DecodeReq{Tx: string(out.Bytes())})
	s.Require().NoError(err)

	// try to decode the txn using legacy rest, it fails.
	res, err := rest.PostRequest(fmt.Sprintf("%s/txs/decode", val.APIAddress), "application/json", bz)
	s.Require().NoError(err)

	var errResp rest.ErrorResponse
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(res, &errResp))
	s.Require().Contains(errResp.Error, errMsg)
}

// TestLegacyRestErrMessages creates two IBC txs, one that fails, one that
// succeeds, and make sure we cannot query any of them (with pretty error msg).
// Our intension is to test the error message of querying a message which is
// signed with proto, since IBC won't support legacy amino at all we are
// considering a message from IBC module.
func (s *IntegrationTestSuite) TestLegacyRestErrMessages() {
	val := s.network.Validators[0]

	// Write consensus json to temp file, used for an IBC message.
	consensusJSON := testutil.WriteToNewTempFile(
		s.T(),
		`{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A/3SXL2ONYaOkxpdR5P8tHTlSlPv1AwQwSFxKRee5JQW"},"diversifier":"diversifier","timestamp":"10"}`,
	)

	testCases := []struct {
		desc string
		cmd  *cobra.Command
		args []string
		code uint32
	}{
		{
			"Failing IBC message",
			ibccli.NewChannelCloseInitCmd(),
			[]string{
				"121",       // dummy port-id
				"channel-0", // dummy channel-id
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=foobar", flags.FlagMemo),
			},
			uint32(7),
		},
		{
			"Successful IBC message",
			ibcsolomachinecli.NewCreateClientCmd(),
			[]string{
				"1",                  // dummy sequence
				consensusJSON.Name(), // path to consensus json,
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=foobar", flags.FlagMemo),
			},
			uint32(0),
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, tc.cmd, tc.args)
			s.Require().NoError(err)
			var txRes sdk.TxResponse
			s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txRes))
			s.Require().Equal(tc.code, txRes.Code)

			s.Require().NoError(s.network.WaitForNextBlock())

			s.testQueryIBCTx(txRes, tc.cmd, tc.args)
		})
	}
}

// TestLegacyMultiSig creates a legacy multisig transaction, and makes sure
// we can query it via the legacy REST endpoint.
// ref: https://github.com/cosmos/cosmos-sdk/issues/8679
func (s *IntegrationTestSuite) TestLegacyMultisig() {
	val1 := *s.network.Validators[0]

	// Generate 2 accounts and a multisig.
	account1, err := val1.ClientCtx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := val1.ClientCtx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigInfo, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 1000)
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

	// Generate multisig transaction to a random address.
	_, _, recipient := testdata.KeyTestPubAddr()
	multiGeneratedTx, err := bankcli.MsgSendExec(
		val1.ClientCtx,
		multisigInfo.GetAddress(),
		recipient,
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

	// Does not work in offline mode.
	_, err = authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), "--offline", sign1File.Name(), sign2File.Name())
	s.Require().EqualError(err, fmt.Sprintf("couldn't verify signature: unable to verify single signer signature"))

	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())

	_, err = authtest.TxValidateSignaturesExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	val1.ClientCtx.BroadcastMode = flags.BroadcastBlock
	out, err := authtest.TxBroadcastExec(val1.ClientCtx, signedTxFile.Name())
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())

	var txRes sdk.TxResponse
	err = val1.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txRes)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), txRes.Code)

	s.testQueryTx(txRes.Height, txRes.TxHash, recipient.String())
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
