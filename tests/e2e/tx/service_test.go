package tx_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtest "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var bankMsgSendEventAction = fmt.Sprintf("message.action='%s'", sdk.MsgTypeURL(&banktypes.MsgSend{}))

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	txHeight    int64
	queryClient tx.ServiceClient
	txRes       sdk.TxResponse
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	s.cfg = cfg

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	val := s.network.Validators[0]
	s.Require().NoError(s.network.WaitForNextBlock())

	s.queryClient = tx.NewServiceClient(val.ClientCtx)

	// Create a new MsgSend tx from val to itself.
	out, err := cli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10)),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
		fmt.Sprintf("--%s=foobar", flags.FlagNote),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &s.txRes))
	s.Require().Equal(uint32(0), s.txRes.Code, s.txRes)

	out, err = cli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, math.NewInt(1)),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s", flags.FlagOffline),
		fmt.Sprintf("--%s=0", flags.FlagAccountNumber),
		fmt.Sprintf("--%s=2", flags.FlagSequence),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
		fmt.Sprintf("--%s=foobar", flags.FlagNote),
	)
	s.Require().NoError(err)
	var tr sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &tr))
	s.Require().Equal(uint32(0), tr.Code)

	resp, err := cli.GetTxResponse(s.network, val.ClientCtx, tr.TxHash)
	s.Require().NoError(err)
	s.txHeight = resp.Height
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestQueryBySig() {
	// broadcast tx
	txb := s.mkTxBuilder()
	txbz, err := s.cfg.TxConfig.TxEncoder()(txb.GetTx())
	s.Require().NoError(err)
	resp, err := s.queryClient.BroadcastTx(context.Background(), &tx.BroadcastTxRequest{TxBytes: txbz, Mode: tx.BroadcastMode_BROADCAST_MODE_SYNC})
	s.Require().NoError(err)
	s.Require().NotEmpty(resp.TxResponse.TxHash)

	s.Require().NoError(s.network.WaitForNextBlock())

	// get the signature out of the builder
	sigs, err := txb.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Len(sigs, 1)
	sig, ok := sigs[0].Data.(*signing.SingleSignatureData)
	s.Require().True(ok)

	// encode, format, query
	b64Sig := base64.StdEncoding.EncodeToString(sig.Signature)
	sigFormatted := fmt.Sprintf("%s.%s='%s'", sdk.EventTypeTx, sdk.AttributeKeySignature, b64Sig)
	res, err := s.queryClient.GetTxsEvent(context.Background(), &tx.GetTxsEventRequest{
		Query:   sigFormatted,
		OrderBy: 0,
		Page:    0,
		Limit:   10,
	})
	s.Require().NoError(err)
	s.Require().Len(res.Txs, 1)
	s.Require().Len(res.Txs[0].Signatures, 1)
	s.Require().Equal(res.Txs[0].Signatures[0], sig.Signature)
}

func (s *E2ETestSuite) TestSimulateTx_GRPC() {
	val := s.network.Validators[0]
	txBuilder := s.mkTxBuilder()
	// Convert the txBuilder to a tx.Tx.
	protoTx, err := txBuilderToProtoTx(txBuilder)
	s.Require().NoError(err)
	// Encode the txBuilder to txBytes.
	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *tx.SimulateRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.SimulateRequest{}, true, "empty txBytes is not allowed"},
		{"valid request with proto tx (deprecated)", &tx.SimulateRequest{Tx: protoTx}, false, ""},
		{"valid request with tx_bytes", &tx.SimulateRequest{TxBytes: txBytes}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Broadcast the tx via gRPC via the validator's clientCtx (which goes
			// through Tendermint).
			res, err := s.queryClient.Simulate(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				// Check the result and gas used are correct.
				//
				// The 12 events are:
				// - Sending Fee to the pool: coin_spent, coin_received, transfer and message.sender=<val1>
				// - tx.* events: tx.fee, tx.acc_seq, tx.signature
				// - Sending Amount to recipient: coin_spent, coin_received, transfer and message.sender=<val1>
				// - Msg events: message.module=bank and message.action=/cosmos.bank.v1beta1.MsgSend (in one message)
				s.Require().Equal(12, len(res.GetResult().GetEvents()))
				s.Require().True(res.GetGasInfo().GetGasUsed() > 0) // Gas used sometimes change, just check it's not empty.
			}
		})
	}
}

func (s *E2ETestSuite) TestSimulateTx_GRPCGateway() {
	val := s.network.Validators[0]
	txBuilder := s.mkTxBuilder()
	// Convert the txBuilder to a tx.Tx.
	protoTx, err := txBuilderToProtoTx(txBuilder)
	s.Require().NoError(err)
	// Encode the txBuilder to txBytes.
	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *tx.SimulateRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.SimulateRequest{}, true, "empty txBytes is not allowed"},
		{"valid request with proto tx (deprecated)", &tx.SimulateRequest{Tx: protoTx}, false, ""},
		{"valid request with tx_bytes", &tx.SimulateRequest{TxBytes: txBytes}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.ClientCtx.Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)
			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/simulate", val.APIAddress), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.SimulateResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				// Check the result and gas used are correct.
				s.Require().Len(result.GetResult().MsgResponses, 1)
				s.Require().Equal(12, len(result.GetResult().GetEvents())) // See TestSimulateTx_GRPC for the 12 events.
				s.Require().True(result.GetGasInfo().GetGasUsed() > 0)     // Gas used sometimes change, just check it's not empty.
			}
		})
	}
}

func (s *E2ETestSuite) TestGetTxEvents_GRPC() {
	testCases := []struct {
		name      string
		req       *tx.GetTxsEventRequest
		expErr    bool
		expErrMsg string
		expLen    int
	}{
		{
			"nil request",
			nil,
			true,
			"request cannot be nil",
			0,
		},
		{
			"empty request",
			&tx.GetTxsEventRequest{},
			true,
			"query cannot be empty",
			0,
		},
		{
			"request with dummy event",
			&tx.GetTxsEventRequest{Query: "foobar"},
			true,
			"failed to search for txs",
			0,
		},
		{
			"request with order-by",
			&tx.GetTxsEventRequest{
				Query:   bankMsgSendEventAction,
				OrderBy: tx.OrderBy_ORDER_BY_ASC,
			},
			false,
			"",
			3,
		},
		{
			"without pagination",
			&tx.GetTxsEventRequest{
				Query: bankMsgSendEventAction,
			},
			false,
			"",
			3,
		},
		{
			"with pagination",
			&tx.GetTxsEventRequest{
				Query: bankMsgSendEventAction,
				Page:  1,
				Limit: 2,
			},
			false,
			"",
			2,
		},
		{
			"with multi events",
			&tx.GetTxsEventRequest{
				Query: fmt.Sprintf("%s AND message.module='bank'", bankMsgSendEventAction),
			},
			false,
			"",
			3,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Query the tx via gRPC.
			grpcRes, err := s.queryClient.GetTxsEvent(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().GreaterOrEqual(len(grpcRes.Txs), 1)
				s.Require().Equal("foobar", grpcRes.Txs[0].Body.Memo)
				s.Require().Equal(tc.expLen, len(grpcRes.Txs))

				// Make sure fields are populated.
				// ref: https://github.com/cosmos/cosmos-sdk/issues/8680
				// ref: https://github.com/cosmos/cosmos-sdk/issues/8681
				s.Require().NotEmpty(grpcRes.TxResponses[0].Timestamp)
				s.Require().Empty(grpcRes.TxResponses[0].RawLog) // logs are empty if the transactions are successful
			}
		})
	}
}

func (s *E2ETestSuite) TestGetTxEvents_GRPCGateway() {
	val := s.network.Validators[0]
	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
		expLen    int
	}{
		{
			"empty params",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", val.APIAddress),
			true,
			"query cannot be empty", 0,
		},
		{
			"without pagination",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s", val.APIAddress, bankMsgSendEventAction),
			false,
			"", 3,
		},
		{
			"with pagination",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&page=%d&limit=%d", val.APIAddress, bankMsgSendEventAction, 1, 2),
			false,
			"", 2,
		},
		{
			"valid request: order by asc",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&query=%s&order_by=ORDER_BY_ASC", val.APIAddress, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"", 3,
		},
		{
			"valid request: order by desc",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&query=%s&order_by=ORDER_BY_DESC", val.APIAddress, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"", 3,
		},
		{
			"invalid request: invalid order by",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&query=%s&order_by=invalid_order", val.APIAddress, bankMsgSendEventAction, "message.module='bank'"),
			true,
			"is not a valid tx.OrderBy", 0,
		},
		{
			"expect pass with multiple-events",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s&query=%s", val.APIAddress, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"", 3,
		},
		{
			"expect pass with escape event",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?query=%s", val.APIAddress, "message.action%3D'/cosmos.bank.v1beta1.MsgSend'"),
			false,
			"", 3,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.GetTxsEventResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err, "failed to unmarshal JSON: %s", res)
				s.Require().GreaterOrEqual(len(result.Txs), 1)
				s.Require().Equal("foobar", result.Txs[0].Body.Memo)
				s.Require().NotZero(result.TxResponses[0].Height)
				s.Require().Equal(tc.expLen, len(result.Txs))
			}
		})
	}
}

func (s *E2ETestSuite) TestGetTx_GRPC() {
	testCases := []struct {
		name      string
		req       *tx.GetTxRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.GetTxRequest{}, true, "tx hash cannot be empty"},
		{"request with dummy hash", &tx.GetTxRequest{Hash: "deadbeef"}, true, "code = NotFound desc = tx not found: deadbeef"},
		{"good request", &tx.GetTxRequest{Hash: s.txRes.TxHash}, false, ""},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Query the tx via gRPC.
			grpcRes, err := s.queryClient.GetTx(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().Equal("foobar", grpcRes.Tx.Body.Memo)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetTx_GRPCGateway() {
	val := s.network.Validators[0]
	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{
			"empty params",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/", val.APIAddress),
			true, "tx hash cannot be empty",
		},
		{
			"dummy hash",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", val.APIAddress, "deadbeef"),
			true, "code = NotFound desc = tx not found: deadbeef",
		},
		{
			"good hash",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", val.APIAddress, s.txRes.TxHash),
			false, "",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.GetTxResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				s.Require().Equal("foobar", result.Tx.Body.Memo)
				s.Require().NotZero(result.TxResponse.Height)

				// Make sure fields are populated.
				// ref: https://github.com/cosmos/cosmos-sdk/issues/8680
				// ref: https://github.com/cosmos/cosmos-sdk/issues/8681
				s.Require().NotEmpty(result.TxResponse.Timestamp)
				s.Require().Empty(result.TxResponse.RawLog) // logs are empty on successful transactions
			}
		})
	}
}

func (s *E2ETestSuite) TestBroadcastTx_GRPC() {
	val := s.network.Validators[0]
	txBuilder := s.mkTxBuilder()
	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *tx.BroadcastTxRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.BroadcastTxRequest{}, true, "invalid empty tx"},
		{"no mode", &tx.BroadcastTxRequest{TxBytes: txBytes}, true, "supported types: sync, async"},
		{"valid request", &tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Broadcast the tx via gRPC via the validator's clientCtx (which goes
			// through Tendermint).
			grpcRes, err := s.queryClient.BroadcastTx(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(uint32(0), grpcRes.TxResponse.Code)
			}
		})
	}
}

func (s *E2ETestSuite) TestBroadcastTx_GRPCGateway() {
	val := s.network.Validators[0]
	txBuilder := s.mkTxBuilder()
	txBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *tx.BroadcastTxRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.BroadcastTxRequest{}, true, "invalid empty tx"},
		{"no mode", &tx.BroadcastTxRequest{TxBytes: txBytes}, true, "supported types: sync, async"},
		{"valid request", &tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.ClientCtx.Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)
			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", val.APIAddress), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.BroadcastTxResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				s.Require().Equal(uint32(0), result.TxResponse.Code, "rawlog", result.TxResponse.RawLog)
			}
		})
	}
}

func (s *E2ETestSuite) TestUnorderedCannotUseSequence() {
	val1 := *s.network.Validators[0]
	coins := sdk.NewInt64Coin(s.cfg.BondDenom, 15)
	_, err := cli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		val1.Address,
		sdk.NewCoins(coins),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
		fmt.Sprintf("--sequence=%d", 15),
		"--unordered",
		fmt.Sprintf("--timeout-duration=%s", "10s"),
	)
	s.Require().ErrorContains(err, "if any flags in the group [unordered sequence] are set none of the others can be; [sequence unordered] were all set")
}

func (s *E2ETestSuite) TestSimMultiSigTx() {
	val1 := *s.network.Validators[0]

	kr := val1.ClientCtx.Keyring

	account1, _, err := kr.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	account2, _, err := kr.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	pub1, err := account1.GetPubKey()
	s.Require().NoError(err)

	pub2, err := account2.GetPubKey()
	s.Require().NoError(err)

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{pub1, pub2})
	_, err = kr.SaveMultisig("multi", multi)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())

	multisigRecord, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	height, err := s.network.LatestHeight()
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(height + 1)
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	coins := sdk.NewInt64Coin(s.cfg.BondDenom, 15)
	_, err = cli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		addr,
		sdk.NewCoins(coins),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)
	s.Require().NoError(err)

	height, err = s.network.LatestHeight()
	s.Require().NoError(err)
	_, err = s.network.WaitForHeight(height + 1)
	s.Require().NoError(err)

	// Generate multisig transaction.
	multiGeneratedTx, err := cli.MsgSendExec(
		val1.ClientCtx,
		addr,
		val1.Address,
		sdk.NewCoins(
			sdk.NewInt64Coin(s.cfg.BondDenom, 5),
		),
		addresscodec.NewBech32Codec("cosmos"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
		fmt.Sprintf("--%s=foobar", flags.FlagNote),
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())

	// Sign with account1
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	val1.ClientCtx.HomeDir = strings.Replace(val1.ClientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtest.TxSignExec(val1.ClientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)
	sign1File := testutil.WriteToNewTempFile(s.T(), account1Signature.String())

	// Sign with account2
	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	account2Signature, err := authtest.TxSignExec(val1.ClientCtx, addr2, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	s.Require().NoError(err)
	sign2File := testutil.WriteToNewTempFile(s.T(), account2Signature.String())

	// multisign tx
	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// convert from protoJSON to protoBinary for sim
	sdkTx, err := val1.ClientCtx.TxConfig.TxJSONDecoder()(multiSigWith2Signatures.Bytes())
	s.Require().NoError(err)
	txBytes, err := val1.ClientCtx.TxConfig.TxEncoder()(sdkTx)
	s.Require().NoError(err)

	// simulate tx
	sim := &tx.SimulateRequest{TxBytes: txBytes}
	res, err := s.queryClient.Simulate(context.Background(), sim)
	s.Require().NoError(err)

	// make sure gas was used
	s.Require().Greater(res.GasInfo.GasUsed, uint64(0))
}

func (s *E2ETestSuite) TestGetBlockWithTxs_GRPC() {
	testCases := []struct {
		name      string
		req       *tx.GetBlockWithTxsRequest
		expErr    bool
		expErrMsg string
		expTxsLen int
	}{
		{"nil request", nil, true, "request cannot be nil", 0},
		{"empty request", &tx.GetBlockWithTxsRequest{}, true, "height must not be less than 1 or greater than the current height", 0},
		{"bad height", &tx.GetBlockWithTxsRequest{Height: 99999999}, true, "height must not be less than 1 or greater than the current height", 0},
		{"bad pagination", &tx.GetBlockWithTxsRequest{Height: s.txHeight, Pagination: &query.PageRequest{Offset: 1000, Limit: 100}}, true, "out of range", 0},
		{"good request", &tx.GetBlockWithTxsRequest{Height: s.txHeight}, false, "", 1},
		{"with pagination request", &tx.GetBlockWithTxsRequest{Height: s.txHeight, Pagination: &query.PageRequest{Offset: 0, Limit: 1}}, false, "", 1},
		{"page all request", &tx.GetBlockWithTxsRequest{Height: s.txHeight, Pagination: &query.PageRequest{Offset: 0, Limit: 100}}, false, "", 1},
		{"block with 0 tx", &tx.GetBlockWithTxsRequest{Height: s.txHeight - 1, Pagination: &query.PageRequest{Offset: 0, Limit: 100}}, false, "", 0},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Query the tx via gRPC.
			grpcRes, err := s.queryClient.GetBlockWithTxs(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				if tc.expTxsLen > 0 {
					s.Require().Equal("foobar", grpcRes.Txs[0].Body.Memo)
				}
				s.Require().Equal(grpcRes.Block.Header.Height, tc.req.Height)
				if tc.req.Pagination != nil {
					s.Require().LessOrEqual(len(grpcRes.Txs), int(tc.req.Pagination.Limit))
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestGetBlockWithTxs_GRPCGateway() {
	val := s.network.Validators[0]
	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{
			"empty params",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/block/0", val.APIAddress),
			true, "height must not be less than 1 or greater than the current height",
		},
		{
			"bad height",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/block/%d", val.APIAddress, 9999999),
			true, "height must not be less than 1 or greater than the current height",
		},
		{
			"good request",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/block/%d", val.APIAddress, s.txHeight),
			false, "",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.GetBlockWithTxsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				s.Require().Equal("foobar", result.Txs[0].Body.Memo)
				s.Require().Equal(result.Block.Header.Height, s.txHeight)
			}
		})
	}
}

func (s *E2ETestSuite) TestTxEncode_GRPC() {
	val := s.network.Validators[0]
	txBuilder := s.mkTxBuilder()
	protoTx, err := txBuilderToProtoTx(txBuilder)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *tx.TxEncodeRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxEncodeRequest{}, true, "invalid empty tx"},
		{"valid tx request", &tx.TxEncodeRequest{Tx: protoTx}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.queryClient.TxEncode(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
				s.Require().Empty(res)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(res.GetTxBytes())

				tx, err := val.ClientCtx.TxConfig.TxDecoder()(res.TxBytes)
				s.Require().NoError(err)
				s.Require().Equal(protoTx.GetMsgs(), tx.GetMsgs())
			}
		})
	}
}

func (s *E2ETestSuite) TestTxEncode_GRPCGateway() {
	val := s.network.Validators[0]
	txBuilder := s.mkTxBuilder()
	protoTx, err := txBuilderToProtoTx(txBuilder)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *tx.TxEncodeRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxEncodeRequest{}, true, "invalid empty tx"},
		{"valid tx request", &tx.TxEncodeRequest{Tx: protoTx}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.ClientCtx.Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/encode", val.APIAddress), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.TxEncodeResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)

				tx, err := val.ClientCtx.TxConfig.TxDecoder()(result.TxBytes)
				s.Require().NoError(err)
				s.Require().Equal(protoTx.GetMsgs(), tx.GetMsgs())
			}
		})
	}
}

func (s *E2ETestSuite) TestTxDecode_GRPC() {
	val := s.network.Validators[0]
	txBuilder := s.mkTxBuilder()

	encodedTx, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	invalidTxBytes := append(encodedTx, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxDecodeRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeRequest{TxBytes: invalidTxBytes}, true, "tx parse error"},
		{"valid request with tx bytes", &tx.TxDecodeRequest{TxBytes: encodedTx}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.queryClient.TxDecode(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
				s.Require().Empty(res)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(res.GetTx())

				txb := authtx.WrapTx(res.Tx)
				tx, err := val.ClientCtx.TxConfig.TxEncoder()(txb.GetTx())
				s.Require().NoError(err)
				s.Require().Equal(encodedTx, tx)
			}
		})
	}
}

func (s *E2ETestSuite) TestTxDecode_GRPCGateway() {
	val := s.network.Validators[0]
	txBuilder := s.mkTxBuilder()

	encodedTxBytes, err := val.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	invalidTxBytes := append(encodedTxBytes, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxDecodeRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeRequest{TxBytes: invalidTxBytes}, true, "tx parse error"},
		{"valid request with tx_bytes", &tx.TxDecodeRequest{TxBytes: encodedTxBytes}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.ClientCtx.Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/decode", val.APIAddress), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.TxDecodeResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)

				txb := authtx.WrapTx(result.Tx)
				tx, err := val.ClientCtx.TxConfig.TxEncoder()(txb.GetTx())
				s.Require().NoError(err)
				s.Require().Equal(encodedTxBytes, tx)
			}
		})
	}
}

func (s *E2ETestSuite) readTestAminoTxJSON() ([]byte, *legacytx.StdTx) { // nolint:staticcheck // legacy testing
	val := s.network.Validators[0]
	txJSONBytes, err := os.ReadFile("testdata/tx_amino1.json")
	s.Require().NoError(err)
	var stdTx legacytx.StdTx // nolint:staticcheck // legacy testing
	err = val.ClientCtx.LegacyAmino.UnmarshalJSON(txJSONBytes, &stdTx)
	s.Require().NoError(err)
	return txJSONBytes, &stdTx
}

func (s *E2ETestSuite) TestTxEncodeAmino_GRPC() {
	val := s.network.Validators[0]
	txJSONBytes, stdTx := s.readTestAminoTxJSON()

	testCases := []struct {
		name      string
		req       *tx.TxEncodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxEncodeAminoRequest{}, true, "invalid empty tx json"},
		{"invalid request", &tx.TxEncodeAminoRequest{AminoJson: "invalid tx json"}, true, "invalid request"},
		{"valid request with amino-json", &tx.TxEncodeAminoRequest{AminoJson: string(txJSONBytes)}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.queryClient.TxEncodeAmino(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
				s.Require().Empty(res)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(res.GetAminoBinary())

				var decodedTx legacytx.StdTx // nolint:staticcheck // legacy testing
				err = val.ClientCtx.LegacyAmino.Unmarshal(res.AminoBinary, &decodedTx)
				s.Require().NoError(err)
				s.Require().Equal(decodedTx.GetMsgs(), stdTx.GetMsgs())
			}
		})
	}
}

func (s *E2ETestSuite) TestTxEncodeAmino_GRPCGateway() {
	val := s.network.Validators[0]
	txJSONBytes, stdTx := s.readTestAminoTxJSON()

	testCases := []struct {
		name      string
		req       *tx.TxEncodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxEncodeAminoRequest{}, true, "invalid empty tx json"},
		{"invalid request", &tx.TxEncodeAminoRequest{AminoJson: "invalid tx json"}, true, "invalid request"},
		{"valid request with amino-json", &tx.TxEncodeAminoRequest{AminoJson: string(txJSONBytes)}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.ClientCtx.Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/encode/amino", val.APIAddress), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.TxEncodeAminoResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)

				var decodedTx legacytx.StdTx // nolint:staticcheck // legacy testing
				err = val.ClientCtx.LegacyAmino.Unmarshal(result.AminoBinary, &decodedTx)
				s.Require().NoError(err)
				s.Require().Equal(decodedTx.GetMsgs(), stdTx.GetMsgs())
			}
		})
	}
}

func (s *E2ETestSuite) readTestAminoTxBinary() ([]byte, *legacytx.StdTx) { // nolint:staticcheck // legacy testing
	val := s.network.Validators[0]
	txJSONBytes, err := os.ReadFile("testdata/tx_amino1.bin")
	s.Require().NoError(err)
	var stdTx legacytx.StdTx // nolint:staticcheck // legacy testing
	err = val.ClientCtx.LegacyAmino.Unmarshal(txJSONBytes, &stdTx)
	s.Require().NoError(err)
	return txJSONBytes, &stdTx
}

func (s *E2ETestSuite) TestTxDecodeAmino_GRPC() {
	encodedTx, stdTx := s.readTestAminoTxBinary()

	invalidTxBytes := append(encodedTx, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"nil request", nil, true, "request cannot be nil"},
		{"empty request", &tx.TxDecodeAminoRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: invalidTxBytes}, true, "invalid request"},
		{"valid request with tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: encodedTx}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.queryClient.TxDecodeAmino(context.Background(), tc.req)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
				s.Require().Empty(res)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(res.GetAminoJson())

				var decodedTx legacytx.StdTx // nolint:staticcheck // legacy testing
				err = s.network.Validators[0].ClientCtx.LegacyAmino.UnmarshalJSON([]byte(res.GetAminoJson()), &decodedTx)
				s.Require().NoError(err)
				s.Require().Equal(stdTx.GetMsgs(), decodedTx.GetMsgs())
			}
		})
	}
}

func (s *E2ETestSuite) TestTxDecodeAmino_GRPCGateway() {
	val := s.network.Validators[0]
	encodedTx, stdTx := s.readTestAminoTxBinary()

	invalidTxBytes := append(encodedTx, byte(0o00))

	testCases := []struct {
		name      string
		req       *tx.TxDecodeAminoRequest
		expErr    bool
		expErrMsg string
	}{
		{"empty request", &tx.TxDecodeAminoRequest{}, true, "invalid empty tx bytes"},
		{"invalid tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: invalidTxBytes}, true, "invalid request"},
		{"valid request with tx bytes", &tx.TxDecodeAminoRequest{AminoBinary: encodedTx}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.ClientCtx.Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)

			res, err := testutil.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/decode/amino", val.APIAddress), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.TxDecodeAminoResponse
				err := val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)

				var decodedTx legacytx.StdTx // nolint:staticcheck // legacy testing
				err = val.ClientCtx.LegacyAmino.UnmarshalJSON([]byte(result.AminoJson), &decodedTx)
				s.Require().NoError(err)
				s.Require().Equal(stdTx.GetMsgs(), decodedTx.GetMsgs())
			}
		})
	}
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) mkTxBuilder() client.TxBuilder {
	val := s.network.Validators[0]
	s.Require().NoError(s.network.WaitForNextBlock())

	// prepare txBuilder with msg
	txBuilder := val.ClientCtx.TxConfig.NewTxBuilder()
	feeAmount := sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 10)}
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(
		txBuilder.SetMsgs(&banktypes.MsgSend{
			FromAddress: val.Address.String(),
			ToAddress:   val.Address.String(),
			Amount:      sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 10)},
		}),
	)
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetMemo("foobar")
	signers, err := txBuilder.GetTx().GetSigners()
	s.Require().NoError(err)
	s.Require().Equal([][]byte{val.Address}, signers)

	// setup txFactory
	txFactory := clienttx.Factory{}.
		WithChainID(val.ClientCtx.ChainID).
		WithKeybase(val.ClientCtx.Keyring).
		WithTxConfig(val.ClientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// Sign Tx.
	err = authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, false, true)
	s.Require().NoError(err)

	return txBuilder
}

// protoTxProvider is a type which can provide a proto transaction. It is a
// workaround to get access to the wrapper TxBuilder's method GetProtoTx().
// Deprecated: It's only used for testing the deprecated Simulate gRPC endpoint
// using a proto Tx field.
type protoTxProvider interface {
	GetProtoTx() *tx.Tx
}

// txBuilderToProtoTx converts a txBuilder into a proto tx.Tx.
// Deprecated: It's used for testing the deprecated Simulate gRPC endpoint
// using a proto Tx field and for testing the TxEncode endpoint.
func txBuilderToProtoTx(txBuilder client.TxBuilder) (*tx.Tx, error) {
	protoProvider, ok := txBuilder.(protoTxProvider)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "expected proto tx builder, got %T", txBuilder)
	}

	return protoProvider.GetProtoTx(), nil
}
