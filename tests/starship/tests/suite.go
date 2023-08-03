package main

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	configFile = "../configs/devnet.yaml"
	chainID    = "simapp"
	denom      = "stake"
)

type TestSuite struct {
	suite.Suite

	config   *Config
	cdc      Codec
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

	//db := dbm.NewMemDB()
	//logger := log.NewTestLogger(s.T())
	//// todo: cannot use simapp, since it starts a server, will have to create txConfig manually
	//app := simapp.NewSimappWithCustomOptions(s.T(), false, simapp.SetupOptions{
	//	Logger:  logger.With("instance", "first"),
	//	DB:      db,
	//	AppOpts: simtestutil.NewAppOptionsWithFlagHome(s.T().TempDir()),
	//})
	//s.app = app
	//s.txConfig = app.TxConfig()

	//tempApp := simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(s.T().TempDir()))
	//encodingConfig := params.EncodingConfig{
	//	InterfaceRegistry: tempApp.InterfaceRegistry(),
	//	Codec:             tempApp.AppCodec(),
	//	TxConfig:          tempApp.TxConfig(),
	//	Amino:             tempApp.LegacyAmino(),
	//}
	//
	//s.encodingConfig = encodingConfig
	//s.txConfig = encodingConfig.TxConfig

	s.cdc = DefaultCodec()

	grpcConn, err := grpc.Dial(
		fmt.Sprintf("0.0.0.0:%d", config.GetChain(chainID).Ports.Grpc),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(s.cdc.InterfaceRegistry).GRPCCodec())))
	s.Require().NoError(err)
	s.grpcConn = grpcConn
}

func (s *TestSuite) TearDownTest() {
	s.T().Log("tearing down e2e integration test suite...")
	err := s.grpcConn.Close()
	s.Require().NoError(err)
}
