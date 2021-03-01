// +build norace

package grpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/tx"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	conn    *grpc.ClientConn
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = network.DefaultConfig()
	s.network = network.New(s.T(), s.cfg)
	s.Require().NotNil(s.network)

	_, err := s.network.WaitForHeight(2)
	s.Require().NoError(err)

	val0 := s.network.Validators[0]
	s.conn, err = grpc.Dial(
		val0.AppConfig.GRPC.Address,
		grpc.WithInsecure(), // Or else we get "no transport security set"
	)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.conn.Close()
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGRPCServer_TestService() {
	// gRPC query to test service should work
	testClient := testdata.NewQueryClient(s.conn)
	testRes, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)
}

func (s *IntegrationTestSuite) TestGRPCServer_BankBalance() {
	val0 := s.network.Validators[0]

	// gRPC query to bank service should work
	denom := fmt.Sprintf("%stoken", val0.Moniker)
	bankClient := banktypes.NewQueryClient(s.conn)
	var header metadata.MD
	bankRes, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom},
		grpc.Header(&header), // Also fetch grpc header
	)
	s.Require().NoError(err)
	s.Require().Equal(
		sdk.NewCoin(denom, s.network.Config.AccountTokens),
		*bankRes.GetBalance(),
	)
	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	s.Require().NotEmpty(blockHeight[0]) // Should contain the block height

	// Request metadata should work
	bankRes, err = bankClient.Balance(
		metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, "1"), // Add metadata to request
		&banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom},
		grpc.Header(&header),
	)
	blockHeight = header.Get(grpctypes.GRPCBlockHeightHeader)
	s.Require().NotEmpty(blockHeight[0]) // blockHeight is []string, first element is block height.
}

func (s *IntegrationTestSuite) TestGRPCServer_Reflection() {
	// Test server reflection
	reflectClient := rpb.NewServerReflectionClient(s.conn)
	stream, err := reflectClient.ServerReflectionInfo(context.Background(), grpc.WaitForReady(true))
	s.Require().NoError(err)
	s.Require().NoError(stream.Send(&rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_ListServices{},
	}))
	res, err := stream.Recv()
	s.Require().NoError(err)
	services := res.GetListServicesResponse().Service
	servicesMap := make(map[string]bool)
	for _, s := range services {
		servicesMap[s.Name] = true
	}
	// Make sure the following services are present
	s.Require().True(servicesMap["cosmos.bank.v1beta1.Query"])
}

func (s *IntegrationTestSuite) TestGRPCServer_GetTxsEvent() {
	// Query the tx via gRPC without pagination. This used to panic, see
	// https://github.com/cosmos/cosmos-sdk/issues/8038.
	txServiceClient := txtypes.NewServiceClient(s.conn)
	_, err := txServiceClient.GetTxsEvent(
		context.Background(),
		&tx.GetTxsEventRequest{
			Events: []string{"message.action='send'"},
		},
	)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestGRPCServer_BroadcastTx() {
	val0 := s.network.Validators[0]

	// prepare txBuilder with msg
	txBuilder := val0.ClientCtx.TxConfig.NewTxBuilder()
	feeAmount := sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 10)}
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(
		txBuilder.SetMsgs(&banktypes.MsgSend{
			FromAddress: val0.Address.String(),
			ToAddress:   val0.Address.String(),
			Amount:      sdk.Coins{sdk.NewInt64Coin(s.cfg.BondDenom, 10)},
		}),
	)
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)

	// setup txFactory
	txFactory := clienttx.Factory{}.
		WithChainID(val0.ClientCtx.ChainID).
		WithKeybase(val0.ClientCtx.Keyring).
		WithTxConfig(val0.ClientCtx.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// Sign Tx.
	err := authclient.SignTx(txFactory, val0.ClientCtx, val0.Moniker, txBuilder, false, true)
	s.Require().NoError(err)

	txBytes, err := val0.ClientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	// Broadcast the tx via gRPC.
	queryClient := tx.NewServiceClient(s.conn)
	grpcRes, err := queryClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), grpcRes.TxResponse.Code)
}

// Test and enforce that we upfront reject any connections to baseapp containing
// invalid initial x-cosmos-block-height that aren't positive  and in the range [0, max(int64)]
// See issue https://github.com/cosmos/cosmos-sdk/issues/7662.
func (s *IntegrationTestSuite) TestGRPCServerInvalidHeaderHeights() {
	t := s.T()
	val0 := s.network.Validators[0]

	// We should reject connections with invalid block heights off the bat.
	invalidHeightStrs := []struct {
		value   string
		wantErr string
	}{
		{"-1", "\"x-cosmos-block-height\" must be >= 0"},
		{"9223372036854775808", "value out of range"}, // > max(int64) by 1
		{"-10", "\"x-cosmos-block-height\" must be >= 0"},
		{"18446744073709551615", "value out of range"}, // max uint64, which is  > max(int64)
		{"-9223372036854775809", "value out of range"}, // Out of the range of for negative int64
	}
	for _, tt := range invalidHeightStrs {
		t.Run(tt.value, func(t *testing.T) {
			conn, err := grpc.Dial(
				val0.AppConfig.GRPC.Address,
				grpc.WithInsecure(), // Or else we get "no transport security set"
			)
			defer conn.Close()

			testClient := testdata.NewQueryClient(conn)
			ctx := metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, tt.value)
			testRes, err := testClient.Echo(ctx, &testdata.EchoRequest{Message: "hello"})
			require.Error(t, err)
			require.Nil(t, testRes)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
