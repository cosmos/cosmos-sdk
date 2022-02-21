// build +norace

package tx_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtest "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var bankMsgSendEventAction = fmt.Sprintf("message.action='%s'", sdk.MsgTypeURL(&banktypes.MsgSend{}))

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	queryClient tx.ServiceClient
	txRes       sdk.TxResponse
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)
	s.Require().NotNil(s.network)

	val := s.network.Validators[0]

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.queryClient = tx.NewServiceClient(val.ClientCtx)

	// Create a new MsgSend tx from val to itself.
	out, err := bankcli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		val.Address,
		sdk.NewCoins(
			sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)),
		),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
		fmt.Sprintf("--%s=foobar", flags.FlagNote),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &s.txRes))
	s.Require().Equal(uint32(0), s.txRes.Code)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s IntegrationTestSuite) TestSimulateTx_GRPC() {
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
		tc := tc
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
				s.Require().Equal(len(res.GetResult().GetEvents()), 6) // 1 coin recv 1 coin spent, 1 transfer, 3 messages.
				s.Require().True(res.GetGasInfo().GetGasUsed() > 0)    // Gas used sometimes change, just check it's not empty.
			}
		})
	}
}

func (s IntegrationTestSuite) TestSimulateTx_GRPCGateway() {
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
			res, err := rest.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/simulate", val.APIAddress), "application/json", req)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.SimulateResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				// Check the result and gas used are correct.
				s.Require().Equal(len(result.GetResult().GetEvents()), 6) // 1 coin recv, 1 coin spent,1 transfer, 3 messages.
				s.Require().True(result.GetGasInfo().GetGasUsed() > 0)    // Gas used sometimes change, just check it's not empty.
			}
		})
	}
}

func (s IntegrationTestSuite) TestGetTxEvents_GRPC() {
	testCases := []struct {
		name      string
		req       *tx.GetTxsEventRequest
		expErr    bool
		expErrMsg string
	}{
		{
			"nil request",
			nil,
			true, "request cannot be nil",
		},
		{
			"empty request",
			&tx.GetTxsEventRequest{},
			true, "must declare at least one event to search",
		},
		{
			"request with dummy event",
			&tx.GetTxsEventRequest{Events: []string{"foobar"}},
			true, "event foobar should be of the format: {eventType}.{eventAttribute}={value}",
		},
		{
			"request with order-by",
			&tx.GetTxsEventRequest{
				Events:  []string{bankMsgSendEventAction},
				OrderBy: tx.OrderBy_ORDER_BY_ASC,
			},
			false, "",
		},
		{
			"without pagination",
			&tx.GetTxsEventRequest{
				Events: []string{bankMsgSendEventAction},
			},
			false, "",
		},
		{
			"with pagination",
			&tx.GetTxsEventRequest{
				Events: []string{bankMsgSendEventAction},
				Pagination: &query.PageRequest{
					CountTotal: false,
					Offset:     0,
					Limit:      1,
				},
			},
			false, "",
		},
		{
			"with multi events",
			&tx.GetTxsEventRequest{
				Events: []string{bankMsgSendEventAction, "message.module='bank'"},
			},
			false, "",
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

				// Make sure fields are populated.
				// ref: https://github.com/cosmos/cosmos-sdk/issues/8680
				// ref: https://github.com/cosmos/cosmos-sdk/issues/8681
				s.Require().NotEmpty(grpcRes.TxResponses[0].Timestamp)
				s.Require().NotEmpty(grpcRes.TxResponses[0].RawLog)
			}
		})
	}
}

func (s IntegrationTestSuite) TestGetTxEvents_GRPCGateway() {
	val := s.network.Validators[0]
	testCases := []struct {
		name      string
		url       string
		expErr    bool
		expErrMsg string
	}{
		{
			"empty params",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", val.APIAddress),
			true,
			"must declare at least one event to search",
		},
		{
			"without pagination",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=%s", val.APIAddress, bankMsgSendEventAction),
			false,
			"",
		},
		{
			"with pagination",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=%s&pagination.offset=%d&pagination.limit=%d", val.APIAddress, bankMsgSendEventAction, 0, 10),
			false,
			"",
		},
		{
			"valid request: order by asc",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=%s&events=%s&order_by=ORDER_BY_ASC", val.APIAddress, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"",
		},
		{
			"valid request: order by desc",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=%s&events=%s&order_by=ORDER_BY_DESC", val.APIAddress, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"",
		},
		{
			"invalid request: invalid order by",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=%s&events=%s&order_by=invalid_order", val.APIAddress, bankMsgSendEventAction, "message.module='bank'"),
			true,
			"is not a valid tx.OrderBy",
		},
		{
			"expect pass with multiple-events",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=%s&events=%s", val.APIAddress, bankMsgSendEventAction, "message.module='bank'"),
			false,
			"",
		},
		{
			"expect pass with escape event",
			fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?events=%s", val.APIAddress, "message.action%3D'/cosmos.bank.v1beta1.MsgSend'"),
			false,
			"",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)
			if tc.expErr {
				s.Require().Contains(string(res), tc.expErrMsg)
			} else {
				var result tx.GetTxsEventResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(res, &result)
				s.Require().NoError(err)
				s.Require().GreaterOrEqual(len(result.Txs), 1)
				s.Require().Equal("foobar", result.Txs[0].Body.Memo)
				s.Require().NotZero(result.TxResponses[0].Height)
			}
		})
	}
}

func (s IntegrationTestSuite) TestGetTx_GRPC() {
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

func (s IntegrationTestSuite) TestGetTx_GRPCGateway() {
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
			res, err := rest.GetRequest(tc.url)
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
				s.Require().NotEmpty(result.TxResponse.RawLog)
			}
		})
	}
}

func (s IntegrationTestSuite) TestBroadcastTx_GRPC() {
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
		{"no mode", &tx.BroadcastTxRequest{
			TxBytes: txBytes,
		}, true, "supported types: sync, async, block"},
		{"valid request", &tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		}, false, ""},
	}

	for _, tc := range testCases {
		tc := tc
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

func (s IntegrationTestSuite) TestBroadcastTx_GRPCGateway() {
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
		{"no mode", &tx.BroadcastTxRequest{TxBytes: txBytes}, true, "supported types: sync, async, block"},
		{"valid request", &tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		}, false, ""},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := val.ClientCtx.Codec.MarshalJSON(tc.req)
			s.Require().NoError(err)
			res, err := rest.PostRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", val.APIAddress), "application/json", req)
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

func (s *IntegrationTestSuite) TestSimMultiSigTx() {
	val1 := *s.network.Validators[0]

	kr := val1.ClientCtx.Keyring

	account1, _, err := kr.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	account2, _, err := kr.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{account1.GetPubKey(), account2.GetPubKey()})
	_, err = kr.SaveMultisig("multi", multi)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	multisigInfo, err := val1.ClientCtx.Keyring.Key("multi")
	s.Require().NoError(err)

	height, err := s.network.LatestHeight()
	_, err = s.network.WaitForHeight(height + 1)
	s.Require().NoError(err)

	// Send coins from validator to multisig.
	coins := sdk.NewInt64Coin(s.cfg.BondDenom, 15)
	_, err = bankcli.MsgSendExec(
		val1.ClientCtx,
		val1.Address,
		multisigInfo.GetAddress(),
		sdk.NewCoins(coins),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--gas=%d", flags.DefaultGasLimit),
	)

	height, err = s.network.LatestHeight()
	_, err = s.network.WaitForHeight(height + 1)
	s.Require().NoError(err)

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
		fmt.Sprintf("--%s=foobar", flags.FlagNote),
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

	// multisign tx
	val1.ClientCtx.Offline = false
	multiSigWith2Signatures, err := authtest.TxMultiSignExec(val1.ClientCtx, multisigInfo.GetName(), multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	s.Require().NoError(err)

	// convert from protoJSON to protoBinary for sim
	sdkTx, err := val1.ClientCtx.TxConfig.TxJSONDecoder()(multiSigWith2Signatures.Bytes())
	txBytes, err := val1.ClientCtx.TxConfig.TxEncoder()(sdkTx)

	// simulate tx
	sim := &tx.SimulateRequest{TxBytes: txBytes}
	res, err := s.queryClient.Simulate(context.Background(), sim)
	s.Require().NoError(err)

	// make sure gas was used
	s.Require().Greater(res.GasInfo.GasUsed, uint64(0))
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s IntegrationTestSuite) mkTxBuilder() client.TxBuilder {
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

	// setup txFactory
	txFactory := clienttx.Factory{}.
		WithChainID(val.ClientCtx.ChainID).
		WithKeybase(val.ClientCtx.Keyring).
		WithTxConfig(val.ClientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// Sign Tx.
	err := authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, false, true)
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
// Deprecated: It's only used for testing the deprecated Simulate gRPC endpoint
// using a proto Tx field.
func txBuilderToProtoTx(txBuilder client.TxBuilder) (*tx.Tx, error) { // nolint
	protoProvider, ok := txBuilder.(protoTxProvider)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected proto tx builder, got %T", txBuilder)
	}

	return protoProvider.GetProtoTx(), nil
}
