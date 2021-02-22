package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (suite *IntegrationTestSuite) TestExportGenesis() {
	app, ctx := suite.app, suite.ctx

	expectedMetadata := suite.getTestMetadata()
	expectedBalances, totalSupply := suite.getTestBalancesAndSupply()
	for i := range []int{1, 2} {
		app.BankKeeper.SetDenomMetaData(ctx, expectedMetadata[i])
		accAddr, err1 := sdk.AccAddressFromBech32(expectedBalances[i].Address)
		if err1 != nil {
			panic(err1)
		}
		// set balances via mint and send
		suite.
			Require().
			NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, expectedBalances[i].Coins))
		suite.
			Require().
			NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, accAddr, expectedBalances[i].Coins))
	}
	app.BankKeeper.SetParams(ctx, types.DefaultParams())

	exportGenesis := app.BankKeeper.ExportGenesis(ctx)

	suite.Require().Len(exportGenesis.Params.SendEnabled, 0)
	suite.Require().Equal(types.DefaultParams().DefaultSendEnabled, exportGenesis.Params.DefaultSendEnabled)
	suite.Require().Equal(totalSupply.Total, exportGenesis.Supply)
	suite.Require().Equal(expectedBalances, exportGenesis.Balances)
	suite.Require().Equal(expectedMetadata, exportGenesis.DenomMetadata)
}

func (suite *IntegrationTestSuite) getTestBalancesAndSupply() ([]types.Balance, *types.Supply) {
	addr2, _ := sdk.AccAddressFromBech32("cosmos1f9xjhxm0plzrh9cskf4qee4pc2xwp0n0556gh0")
	addr1, _ := sdk.AccAddressFromBech32("cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh")
	addr1Balance := sdk.Coins{sdk.NewInt64Coin("testcoin3", 10)}
	addr2Balance := sdk.Coins{sdk.NewInt64Coin("testcoin1", 32), sdk.NewInt64Coin("testcoin2", 34)}

	totalSupply := types.NewSupply(addr1Balance)
	totalSupply.Inflate(addr2Balance)
	return []types.Balance{
		{Address: addr2.String(), Coins: addr2Balance},
		{Address: addr1.String(), Coins: addr1Balance},
	}, totalSupply

}

func (suite *IntegrationTestSuite) TestInitGenesis() {
	m := types.Metadata{Description: sdk.DefaultBondDenom, Base: sdk.DefaultBondDenom, Display: sdk.DefaultBondDenom}
	g := types.DefaultGenesisState()
	g.DenomMetadata = []types.Metadata{m}
	bk := suite.app.BankKeeper
	bk.InitGenesis(suite.ctx, g)

	m2, found := bk.GetDenomMetaData(suite.ctx, m.Base)
	suite.Require().True(found)
	suite.Require().Equal(m, m2)
}
