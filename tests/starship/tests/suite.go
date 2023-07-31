package main

import (
	"os"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client"
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
	app      *simapp.SimApp
	txConfig client.TxConfig
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

	db := dbm.NewMemDB()
	logger := log.NewTestLogger(s.T())
	app := simapp.NewSimappWithCustomOptions(s.T(), false, simapp.SetupOptions{
		Logger:  logger.With("instance", "first"),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(s.T().TempDir()),
	})
	s.app = app
	s.txConfig = app.TxConfig()

	grpcConn, err := grpc.Dial(
		config.GetChain(chainID).GetGRPCAddr(),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())))
	s.Require().NoError(err)
	s.grpcConn = grpcConn
}

func (s *TestSuite) TearDownTest() {
	s.T().Log("tearing down e2e integration test suite...")
	err := s.grpcConn.Close()
	if err != nil {
		panic(err)
	}
}
