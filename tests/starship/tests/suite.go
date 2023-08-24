package main

import (
	"context"
	"fmt"
	"os"
	"time"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	"cosmossdk.io/simapp/params"

	"github.com/cosmos/cosmos-sdk/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

var (
	configFile = "../configs/devnet.yaml"
	chainID    = "simapp"
	denom      = "stake"
)

type TestSuite struct {
	suite.Suite

	config   *Config
	cdc      params.EncodingConfig
	grpcConn *grpc.ClientConn
}

func (s *TestSuite) SetupTest() {
	s.T().Log("setting up e2e integration test suite...")

	// read config file from yaml
	yamlFile, err := os.ReadFile(configFile)
	s.Require().NoError(err)
	config := &Config{}
	err = yaml.Unmarshal(yamlFile, config)
	s.Require().NoError(err)
	s.config = config

	tempApp := simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(s.T().TempDir()))
	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Codec:             tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}

	s.cdc = encodingConfig

	grpcConn, err := grpc.Dial(
		fmt.Sprintf("127.0.0.1:%d", config.GetChain(chainID).Ports.Grpc),
		grpc.WithInsecure(), //nolint:staticcheck // ignore SA1019, we don't need to use a secure connection for tests
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(s.cdc.InterfaceRegistry).GRPCCodec())))
	s.Require().NoError(err)
	s.grpcConn = grpcConn
}

func (s *TestSuite) TearDownTest() {
	s.T().Log("tearing down e2e integration test suite...")
	err := s.grpcConn.Close()
	s.Require().NoError(err)
}

// WaitForTx will wait for the tx to complete, fail if not able to find tx
func (s *TestSuite) WaitForTx(txHex string) {
	var resTx *txtypes.GetTxResponse
	var err error

	txClient := txtypes.NewServiceClient(s.grpcConn)

	s.Require().Eventuallyf(
		func() bool {
			resTx, err = txClient.GetTx(context.Background(), &txtypes.GetTxRequest{Hash: txHex})
			if err != nil {
				return false
			}
			if resTx.TxResponse.Height > 1 {
				return true
			}
			return false
		},
		5*time.Second,
		time.Second,
		"waited for too long, still txn not successful",
	)
	s.Require().NotNil(resTx.TxResponse)
}
