package tx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	query "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	queryClient tx.ServiceClient
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	s.Require().NotNil(s.network)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.queryClient = tx.NewServiceClient(s.network.Validators[0].ClientCtx)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s IntegrationTestSuite) TestSimulate() {
	val := s.network.Validators[0]

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
	err := authclient.SignTx(txFactory, val.ClientCtx, val.Moniker, txBuilder, false)
	s.Require().NoError(err)

	// Convert the txBuilder to a tx.Tx.
	protoTx, err := txBuilderToProtoTx(txBuilder)
	s.Require().NoError(err)

	// Run the simulate gRPC query.
	res, err := s.queryClient.Simulate(
		context.Background(),
		&tx.SimulateRequest{Tx: protoTx},
	)
	s.Require().NoError(err)

	// Check the result and gas used are correct.
	s.Require().Equal(len(res.GetResult().GetEvents()), 4) // 1 transfer, 3 messages.
	s.Require().True(res.GetGasInfo().GetGasUsed() > 0)    // Gas used sometimes change, just check it's not empty.
}

func (s IntegrationTestSuite) TestGetTxEvents() {
	val := s.network.Validators[0]

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
		fmt.Sprintf("--%s=foobar", flags.FlagMemo),
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().Equal(uint32(0), txRes.Code)

	s.Require().NoError(s.network.WaitForNextBlock())

	// Query the tx via gRPC.
	grpcRes, err := s.queryClient.GetTxsEvent(
		context.Background(),
		&tx.GetTxsEventRequest{Event: "message.action=send",
			Pagination: &query.PageRequest{
				CountTotal: false,
				Offset:     0,
				Limit:      1,
			},
		},
	)
	s.Require().NoError(err)
	s.Require().Equal(len(grpcRes.Txs), 1)
	s.Require().Equal("foobar", grpcRes.Txs[0].Body.Memo)

	// Query the tx via grpc-gateway.
	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs?event=%s&pagination.offset=%d&pagination.limit=%d", val.APIAddress, "message.action=send", 0, 1))
	s.Require().NoError(err)
	var getTxRes tx.GetTxsEventResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &getTxRes))
	s.Require().Equal(len(grpcRes.Txs), 1)
	s.Require().Equal("foobar", getTxRes.Txs[0].Body.Memo)
	s.Require().NotZero(grpcRes.TxResponses[0].Height)
}

func (s IntegrationTestSuite) TestGetTx() {
	val := s.network.Validators[0]

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
		fmt.Sprintf("--%s=foobar", flags.FlagMemo),
	)
	s.Require().NoError(err)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txRes))
	s.Require().Equal(uint32(0), txRes.Code)

	s.Require().NoError(s.network.WaitForNextBlock())

	// Query the tx via gRPC.
	grpcRes, err := s.queryClient.GetTx(
		context.Background(),
		&tx.GetTxRequest{Hash: txRes.TxHash},
	)
	s.Require().NoError(err)
	s.Require().Equal("foobar", grpcRes.Tx.Body.Memo)
	s.Require().NotZero(grpcRes.TxResponse.Height)

	// Query the tx via grpc-gateway.
	restRes, err := rest.GetRequest(fmt.Sprintf("%s/cosmos/tx/v1beta1/tx/%s", val.APIAddress, txRes.TxHash))
	s.Require().NoError(err)
	var getTxRes tx.GetTxResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(restRes, &getTxRes))
	s.Require().Equal("foobar", grpcRes.Tx.Body.Memo)
	s.Require().NotZero(grpcRes.TxResponse.Height)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// txBuilderToProtoTx converts a txBuilder into a proto tx.Tx.
func txBuilderToProtoTx(txBuilder client.TxBuilder) (*tx.Tx, error) { // nolint
	intoAnyTx, ok := txBuilder.(codectypes.IntoAny)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (codectypes.IntoAny)(nil), intoAnyTx)
	}

	any := intoAnyTx.AsAny().GetCachedValue()
	if any == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "any's cached value is empty")
	}

	protoTx, ok := any.(*tx.Tx)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (codectypes.IntoAny)(nil), intoAnyTx)
	}

	return protoTx, nil
}
