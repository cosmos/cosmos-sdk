package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	keep "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (suite *KeeperTestSuite) TestQueryAccount() {
	ctx := suite.ctx
	legacyQuerierCdc := codec.NewAminoCodec(suite.encCfg.Amino)

	req := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	path := []string{types.QueryAccount}
	querier := keep.NewQuerier(suite.accountKeeper, legacyQuerierCdc.LegacyAmino)

	bz, err := querier(ctx, []string{"other"}, req)
	suite.Require().Error(err)
	suite.Require().Nil(bz)

	req = abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAccount),
		Data: []byte{},
	}
	res, err := querier(ctx, path, req)
	suite.Require().Error(err)
	suite.Require().Nil(res)

	req.Data = legacyQuerierCdc.MustMarshalJSON(&types.QueryAccountRequest{Address: ""})
	res, err = querier(ctx, path, req)
	suite.Require().Error(err)
	suite.Require().Nil(res)

	_, _, addr := testdata.KeyTestPubAddr()
	req.Data = legacyQuerierCdc.MustMarshalJSON(&types.QueryAccountRequest{Address: addr.String()})
	res, err = querier(ctx, path, req)
	suite.Require().Error(err)
	suite.Require().Nil(res)

	suite.accountKeeper.SetAccount(ctx, suite.accountKeeper.NewAccountWithAddress(ctx, addr))
	res, err = querier(ctx, path, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	res, err = querier(ctx, path, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var account types.AccountI
	err2 := legacyQuerierCdc.LegacyAmino.UnmarshalJSON(res, &account)
	suite.Require().Nil(err2)
}
