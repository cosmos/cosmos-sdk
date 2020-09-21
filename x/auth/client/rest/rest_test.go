package rest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	rest2 "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
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

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestEncodeDecode() {
	val := s.network.Validators[0]

	// NOTE: this uses StdTx explicitly, don't migrate it!
	stdTx := authtypes.StdTx{
		Msgs: []sdk.Msg{&types.MsgSend{}},
		Fee: authtypes.StdFee{
			Amount: sdk.Coins{sdk.NewInt64Coin("foo", 10)},
			Gas:    10000,
		},
		Memo: "FOOBAR",
	}

	// NOTE: this uses amino explicitly, don't migrate it!
	cdc := val.ClientCtx.LegacyAmino

	bz, err := cdc.MarshalJSON(stdTx)
	s.Require().NoError(err)

	res, err := rest.PostRequest(fmt.Sprintf("%s/txs/encode", val.APIAddress), "application/json", bz)
	s.Require().NoError(err)

	var encodeResp rest2.EncodeResp
	err = cdc.UnmarshalJSON(res, &encodeResp)
	s.Require().NoError(err)

	bz, err = cdc.MarshalJSON(rest2.DecodeReq{Tx: encodeResp.Tx})
	s.Require().NoError(err)

	res, err = rest.PostRequest(fmt.Sprintf("%s/txs/decode", val.APIAddress), "application/json", bz)
	s.Require().NoError(err)

	var respWithHeight rest.ResponseWithHeight
	err = cdc.UnmarshalJSON(res, &respWithHeight)
	s.Require().NoError(err)
	var decodeResp rest2.DecodeResp
	err = cdc.UnmarshalJSON(respWithHeight.Result, &decodeResp)
	s.Require().NoError(err)
	s.Require().Equal(stdTx, authtypes.StdTx(decodeResp))
}

func (s *IntegrationTestSuite) TestBroadcastTxRequest() {
	// NOTE: this uses StdTx explicitly, don't migrate it!
	stdTx := authtypes.StdTx{
		Msgs: []sdk.Msg{&types.MsgSend{}},
		Fee: authtypes.StdFee{
			Amount: sdk.Coins{sdk.NewInt64Coin("foo", 10)},
			Gas:    10000,
		},
		Memo: "FOOBAR",
	}

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
	// val1 := s.network.Validators[1]

	// Create and broadcast a tx.
	stdTx := s.createTestStdTx(val0, 1) // Validator's sequence starts at 1.
	res, err := s.broadcastReq(stdTx, "sync")
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	// NOTE: this uses amino explicitly, don't migrate it!
	s.Require().NoError(s.cfg.LegacyAmino.UnmarshalJSON(res, &txRes))
	// we just check for a non-empty TxHash here, the actual hash will depend on the underlying tx configuration
	s.Require().NotEmpty(txRes.TxHash)

	s.network.WaitForNextBlock()
	s.network.WaitForNextBlock()

	// We now fetch the tx by has on the `/tx/{hash}` route.
	txJSON, err := rest.GetRequest(fmt.Sprintf("%s/txs/%s", val0.APIAddress, txRes.TxHash))
	s.Require().NoError(err)

	// txJSON should contain the whole tx, we just make sure that our custom
	// memo is there.
	s.Require().True(strings.Contains(string(txJSON), stdTx.Memo))
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
			stdTx := s.createTestStdTx(s.network.Validators[0], tc.sequence)
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

func (s *IntegrationTestSuite) createTestStdTx(val *network.Validator, sequence uint64) authtypes.StdTx {
	txConfig := authtypes.StdTxConfig{Cdc: s.cfg.LegacyAmino}

	msg := &types.MsgSend{
		FromAddress: val.Address,
		ToAddress:   val.Address,
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
		WithSequence(sequence)

	// sign Tx (offline mode so we can manually set sequence number)
	err := authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, true)
	s.Require().NoError(err)

	stdTx := txBuilder.GetTx().(authtypes.StdTx)

	return stdTx
}

func (s *IntegrationTestSuite) broadcastReq(stdTx authtypes.StdTx, mode string) ([]byte, error) {
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
