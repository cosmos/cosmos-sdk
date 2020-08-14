package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc     *codec.LegacyAmino
	ctx     sdk.Context
	keeper  *keeper.Keeper
	querier sdk.Querier
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.LegacyAmino()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.keeper = app.IBCKeeper
	suite.querier = keeper.NewQuerier(*app.IBCKeeper, app.LegacyAmino())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
