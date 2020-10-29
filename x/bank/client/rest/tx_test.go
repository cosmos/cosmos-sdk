// +build norace

package rest_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankrest "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *IntegrationTestSuite) TestCoinSend() {
	encodingConfig := simapp.MakeTestEncodingConfig()
	authclient.Codec = encodingConfig.Marshaler

	val := s.network.Validators[0]

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
			FromAddress: account.GetAddress().String(),
			ToAddress:   account.GetAddress().String(),
			Amount:      sendReq.Amount,
		},
	}, stdTx.GetMsgs())
}

func submitSendReq(val *network.Validator, req bankrest.SendReq) (legacytx.StdTx, error) {
	url := fmt.Sprintf("%s/bank/accounts/%s/transfers", val.APIAddress, val.Address)

	// NOTE: this uses amino explicitly, don't migrate it!
	bz, err := val.ClientCtx.LegacyAmino.MarshalJSON(req)
	if err != nil {
		return legacytx.StdTx{}, errors.Wrap(err, "error encoding SendReq to json")
	}

	res, err := rest.PostRequest(url, "application/json", bz)
	if err != nil {
		return legacytx.StdTx{}, err
	}

	var tx legacytx.StdTx
	// NOTE: this uses amino explicitly, don't migrate it!
	err = val.ClientCtx.LegacyAmino.UnmarshalJSON(res, &tx)
	if err != nil {
		return legacytx.StdTx{}, errors.Wrap(err, "error unmarshaling to StdTx SendReq response")
	}

	return tx, nil
}

func generateSendReq(from authtypes.AccountI, amount types.Coins) bankrest.SendReq {
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

	return bankrest.SendReq{
		BaseReq: baseReq,
		Amount:  amount,
	}
}

func getAccountInfo(val *network.Validator) (authtypes.AccountI, error) {
	url := fmt.Sprintf("%s/auth/accounts/%s", val.APIAddress, val.Address)

	resp, err := rest.GetRequest(url)
	if err != nil {
		return nil, err
	}

	bz, err := rest.ParseResponseWithHeight(val.ClientCtx.LegacyAmino, resp)
	if err != nil {
		return nil, err
	}

	var acc authtypes.AccountI
	err = val.ClientCtx.LegacyAmino.UnmarshalJSON(bz, &acc)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
