package rest_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
	rest2 "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
)

func (s *IntegrationTestSuite) TestCoinSend() {
	val := s.network.Validators[0]

	initValidatorCoins, err := getCoinsFromValidator(val)
	s.Require().NoError(err)
	s.Require().Equal(
		types.NewCoins(
			types.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
			types.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
		),
		initValidatorCoins,
	)

	account, err := getAccount(val)
	s.Require().NoError(err)

	accnum := account.GetAccountNumber()
	sequence := account.GetSequence()

	baseReq := rest.NewBaseReq(
		account.GetAddress().String(), "someMemo", "some-id", "10000", fmt.Sprintf("%f", 1.0), accnum, sequence, types.NewCoins(), nil, false,
	)

	sendRequest := rest2.SendReq{
		BaseReq: baseReq,
		Amount:  types.Coins{types.NewCoin(s.cfg.BondDenom, types.TokensFromConsensusPower(1))},
	}

	req, err := val.ClientCtx.JSONMarshaler.MarshalJSON(sendRequest)
	s.Require().NoError(err)

	url := fmt.Sprintf("%s/bank/accounts/%s/transfers", val.APIAddress, val.Address)
	resp, err := http.Post(url, "", bytes.NewBuffer(req))
	s.Require().NoError(err)

	bz, err := ioutil.ReadAll(resp.Body)
	s.Require().NoError(err)

	var tx types2.StdTx
	err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(bz, &tx)
	s.Require().NoError(err)
}

func genTx() {

}

func getAccount(val *testutil.Validator) (types2.AccountI, error) {
	url := fmt.Sprintf("%s/auth/accounts/%s", val.APIAddress, val.Address)

	resp, err := rest.GetRequest(url)
	if err != nil {
		return nil, err
	}

	bz, err := rest.ParseResponseWithHeight(val.ClientCtx.JSONMarshaler, resp)
	if err != nil {
		return nil, err
	}

	var acc types2.AccountI
	err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(bz, &acc)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func getCoinsFromValidator(val *testutil.Validator) (types.Coins, error) {
	url := fmt.Sprintf("%s/bank/balances/%s", val.APIAddress, val.Address)

	resp, err := rest.GetRequest(url)
	if err != nil {
		return nil, err
	}

	bz, err := rest.ParseResponseWithHeight(val.ClientCtx.JSONMarshaler, resp)
	var coins types.Coins
	err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(bz, &coins)
	if err != nil {
		return nil, err
	}

	return coins, nil
}
