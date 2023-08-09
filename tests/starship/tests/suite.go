package main

import (
	"fmt"
	"os"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	"cosmossdk.io/simapp/params"

	"github.com/cosmos/cosmos-sdk/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
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
		fmt.Sprintf("0.0.0.0:%d", config.GetChain(chainID).Ports.Grpc),
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
