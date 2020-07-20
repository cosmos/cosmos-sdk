package cli_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	network *testnet.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testnet.DefaultConfig()
	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	var mintData minttypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	inflation := sdk.MustNewDecFromStr("1.0")
	mintData.Minter.Inflation = inflation
	mintData.Params.InflationMin = inflation
	mintData.Params.InflationMax = inflation

	mintDataBz, err := cfg.Codec.MarshalJSON(mintData)
	s.Require().NoError(err)
	genesisState[minttypes.ModuleName] = mintDataBz
	cfg.GenesisState = genesisState

	s.cfg = cfg
	s.network = testnet.New(s.T(), cfg)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetCmdQueryParams() {

}

func (s *IntegrationTestSuite) TestGetCmdQueryInflation() {

}

func (s *IntegrationTestSuite) TestGetCmdQueryAnnualProvisions() {

}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
