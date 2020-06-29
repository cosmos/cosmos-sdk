package rest_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
	rest2 "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

	account, err := getAccountInfo(val)
	s.Require().NoError(err)

	sendReq := generateSendReq(
		account,
		types.Coins{types.NewCoin(s.cfg.BondDenom, types.TokensFromConsensusPower(1))},
	)

	stdTx, err := submitSendReq(val, sendReq)
	s.Require().NoError(err)

	s.Require().Nil(stdTx.Signatures)
	s.Require().Equal([]types.Msg{
		&banktypes.MsgSend{
			FromAddress: account.GetAddress(),
			ToAddress:   account.GetAddress(),
			Amount:      sendReq.Amount,
		},
	}, stdTx.GetMsgs())
}

func submitSendReq(val *testutil.Validator, req rest2.SendReq) (types2.StdTx, error) {
	url := fmt.Sprintf("%s/bank/accounts/%s/transfers", val.APIAddress, val.Address)

	bz, err := val.ClientCtx.JSONMarshaler.MarshalJSON(req)
	if err != nil {
		return types2.StdTx{}, errors.Wrap(err, "error encoding SendReq to json")
	}

	resp, err := http.Post(url, "", bytes.NewBuffer(bz))
	if err != nil {
		return types2.StdTx{}, errors.Wrap(err, "error while sending post request")
	}

	bz, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return types2.StdTx{}, errors.Wrap(err, "error reading SendReq response body")
	}

	var tx types2.StdTx
	err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(bz, &tx)
	if err != nil {
		return types2.StdTx{}, errors.Wrap(err, "error unmarshaling to StdTx SendReq response")
	}

	return tx, nil
}

func generateSendReq(from types2.AccountI, amount types.Coins) rest2.SendReq {
	baseReq := rest.NewBaseReq(
		from.GetAddress().String(),
		"someMemo",
		"some-id",
		"10000",
		fmt.Sprintf("%f", 1.0),
		from.GetAccountNumber(),
		from.GetSequence(),
		types.NewCoins(),
		nil,
		false,
	)

	return rest2.SendReq{
		BaseReq: baseReq,
		Amount:  amount,
	}
}

func getAccountInfo(val *testutil.Validator) (types2.AccountI, error) {
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
