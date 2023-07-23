package main

import (
	"context"
	"fmt"
	"os"

	starship "github.com/cosmology-tech/starship/clients/go/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var (
	configFile = "../configs/devnet.yaml"
	ChainID    = "simapp"
)

type TestSuite struct {
	suite.Suite

	config       *starship.Config
	chainClients starship.ChainClients
}

func (s *TestSuite) SetupTest() {
	s.T().Log("setting up e2e integration test suite...")

	// read config file from yaml
	yamlFile, err := os.ReadFile(configFile)
	s.Require().NoError(err)
	config := &starship.Config{}
	err = yaml.Unmarshal(yamlFile, config)
	s.Require().NoError(err)
	s.config = config

	chainClients, err := starship.NewChainClients(zap.L(), config, nil)
	s.Require().NoError(err)
	s.chainClients = chainClients
}

func (s *TestSuite) TransferTokens(chain *starship.ChainClient, addr string, amount int, denom string) {
	coin, err := sdk.ParseCoinNormalized(fmt.Sprintf("%d%s", amount, denom))
	s.Require().NoError(err)

	// Build transaction message
	req := &banktypes.MsgSend{
		FromAddress: chain.Address,
		ToAddress:   addr,
		Amount:      sdk.Coins{coin},
	}

	res, err := chain.Client.SendMsg(context.Background(), req, "Transfer tokens for e2e tests")
	s.Require().NoError(err)
	s.Require().NotEmpty(res)
}
