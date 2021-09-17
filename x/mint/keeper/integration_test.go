package keeper_test

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
)

type MintKeeperTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
}

func (suite *MintKeeperTestSuite) SetupTest() {
	app := simapp.Setup(true)
	ctx := app.BaseApp.NewContext(true, tmproto.Header{})

	app.MintKeeper.SetParams(ctx, types.DefaultParams())
	app.MintKeeper.SetMinter(ctx, types.DefaultInitialMinter())

	suite.app = app
	suite.ctx = ctx

}
