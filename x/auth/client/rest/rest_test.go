package rest_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"

	rest2 "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
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

func mkTx() legacytx.StdTx {
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

func (s *IntegrationTestSuite) TestEncodeDecode() {
	var require = s.Require()
	val := s.network.Validators[0]
	stdTx := mkTx()

	// NOTE: this uses amino explicitly, don't migrate it!
	cdc := val.ClientCtx.LegacyAmino

	bz, err := cdc.MarshalJSON(stdTx)
	require.NoError(err)

	res, err := rest.PostRequest(fmt.Sprintf("%s/txs/encode", val.APIAddress), "application/json", bz)
	require.NoError(err)

	var encodeResp rest2.EncodeResp
	err = cdc.UnmarshalJSON(res, &encodeResp)
	require.NoError(err)

	bz, err = cdc.MarshalJSON(rest2.DecodeReq{Tx: encodeResp.Tx})
	require.NoError(err)

	res, err = rest.PostRequest(fmt.Sprintf("%s/txs/decode", val.APIAddress), "application/json", bz)
	require.NoError(err)

	var respWithHeight rest.ResponseWithHeight
	err = cdc.UnmarshalJSON(res, &respWithHeight)
	require.NoError(err)
	var decodeResp rest2.DecodeResp
	err = cdc.UnmarshalJSON(respWithHeight.Result, &decodeResp)
	require.NoError(err)
	require.Equal(stdTx, legacytx.StdTx(decodeResp))
}

func (s *IntegrationTestSuite) TestBroadcastTxRequest() {
	stdTx := mkTx()

	// we just test with async mode because this tx will fail - all we care about is that it got encoded and broadcast correctly
	res, err := s.broadcastReq(stdTx, "async")
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	// NOTE: this uses amino explicitly, don't migrate it!
	s.Require().NoError(s.cfg.LegacyAmino.UnmarshalJSON(res, &txRes))
	// we just check for a non-empty TxHash here, the actual hash will depend on the underlying tx configuration
	s.Require().NotEmpty(txRes.TxHash)
}

func (s *IntegrationTestSuite) TestQueryTxByHash() {
	val0 := s.network.Validators[0]

	// We broadcasted a StdTx in SetupSuite.
	// we just check for a non-empty TxHash here, the actual hash will depend on the underlying tx configuration
	s.Require().NotEmpty(s.stdTxRes.TxHash)

	// We now fetch the tx by hash on the `/tx/{hash}` route.
	txJSON, err := rest.GetRequest(fmt.Sprintf("%s/txs/%s", val0.APIAddress, s.stdTxRes.TxHash))
	s.Require().NoError(err)

	// txJSON should contain the whole tx, we just make sure that our custom
	// memo is there.
	s.Require().Contains(string(txJSON), s.stdTx.Memo)
}

func (s *IntegrationTestSuite) TestQueryTxByHeight() {
	val0 := s.network.Validators[0]

	// We broadcasted a StdTx in SetupSuite.
	// we just check for a non-empty height here, as we'll need to for querying.
	s.Require().NotEmpty(s.stdTxRes.Height)

	// We now fetch the tx on `/txs` route, filtering by `tx.height`
	txJSON, err := rest.GetRequest(fmt.Sprintf("%s/txs?limit=10&page=1&tx.height=%d", val0.APIAddress, s.stdTxRes.Height))
	s.Require().NoError(err)

	// txJSON should contain the whole tx, we just make sure that our custom
	// memo is there.
	s.Require().Contains(string(txJSON), s.stdTx.Memo)
}

func (s *IntegrationTestSuite) TestQueryTxByHashWithServiceMessage() {
	val := s.network.Validators[0]

	sendTokens := sdk.NewInt64Coin(s.cfg.BondDenom, 10)

	// Right after this line, we're sending a tx. Might need to wait a block
	// to refresh sequences.
	s.Require().NoError(s.network.WaitForNextBlock())

	out, err := bankcli.ServiceMsgSendExec(
		val.ClientCtx,
		val.Address,
		val.Address,
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

	txJSON, err := rest.GetRequest(fmt.Sprintf("%s/txs/%s", val.APIAddress, txRes.TxHash))
	s.Require().NoError(err)

	var txResAmino sdk.TxResponse
	s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(txJSON, &txResAmino))
	stdTx, ok := txResAmino.Tx.GetCachedValue().(legacytx.StdTx)
	s.Require().True(ok)
	msgs := stdTx.GetMsgs()
	s.Require().Equal(len(msgs), 1)
	_, ok = msgs[0].(*types.MsgSend)
	s.Require().True(ok)
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
	err := authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, true)
	s.Require().NoError(err)

	stdTx := txBuilder.GetTx().(legacytx.StdTx)

	return stdTx
}

func (s *IntegrationTestSuite) broadcastReq(stdTx legacytx.StdTx, mode string) ([]byte, error) {
	val := s.network.Validators[0]

	// NOTE: this uses amino explicitly, don't migrate it!
	cdc := val.ClientCtx.LegacyAmino
	req := rest2.BroadcastReq{
		Tx:   stdTx,
		Mode: mode,
	}
	bz, err := cdc.MarshalJSON(req)
	s.Require().NoError(err)

	return rest.PostRequest(fmt.Sprintf("%s/txs", val.APIAddress), "application/json", bz)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
