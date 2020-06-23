package rest_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type IntegrationTestSuite struct {
	suite.Suite
	cfg     testutil.Config
	network *testutil.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = testutil.NewTestNetwork(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryBalancesRequestHandlerFn() {
	val := s.network.Validators[0]
	apiAddr := val.AppConfig.API.Address

	url := fmt.Sprintf("http://%s/bank/balances/%s?height=1", apiAddr, val.Address)
	resp, err := getRequest(url)
	s.Require().NoError(err)

	bz, err := rest.ParseResponseWithHeight(val.ClientCtx.JSONMarshaler, resp)
	s.Require().NoError(err)

	var balances sdk.Coins
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(bz, &balances))

	expected := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
		sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
	)
	s.Require().Equal(expected.String(), balances.String())

	url = fmt.Sprintf("http://%s/bank/balances/%s?height=1&denom=%s", apiAddr, val.Address, s.cfg.BondDenom)
	resp, err = getRequest(url)
	s.Require().NoError(err)

	bz, err = rest.ParseResponseWithHeight(val.ClientCtx.JSONMarshaler, resp)
	s.Require().NoError(err)

	var balance sdk.Coin
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(bz, &balance))
	s.Require().Equal(expected[1].String(), balance.String())
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func getRequest(url string) ([]byte, error) {
	res, err := http.Get(url) // nolint:gosec
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if err = res.Body.Close(); err != nil {
		return nil, err
	}

	return body, nil
}
