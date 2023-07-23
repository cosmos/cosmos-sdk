package main

import (
	"context"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestChainsStatus() {
	s.T().Log("runing test for /status endpoint for each chain")

	for _, chainClient := range s.chainClients {
		status, err := chainClient.GetStatus()
		s.Require().NoError(err)

		s.Require().Equal(chainClient.GetChainID(), status.NodeInfo.Network)
	}
}

func (s *TestSuite) TestChainTokenTransfer() {
	chain1, err := s.chainClients.GetChainClient(ChainID)
	s.Require().NoError(err)

	keyName := "test-transfer"
	address, err := chain1.CreateRandWallet(keyName)
	s.Require().NoError(err)

	denom, err := chain1.GetChainDenom()
	s.Require().NoError(err)

	s.TransferTokens(chain1, address, 2345000, denom)

	// Verify the address recived the token
	balances, err := banktypes.NewQueryClient(chain1.Client).AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address:    address,
		Pagination: nil,
	})
	s.Require().NoError(err)

	// Assert correct transfers
	s.Require().Len(balances.Balances, 1)
	s.Require().Equal(balances.Balances.Denoms(), []string{denom})
	s.Require().Equal(balances.Balances[0].Amount, sdk.NewInt(2345000))
}
