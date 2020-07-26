package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/KiraCore/cosmos-sdk/codec"
	"github.com/KiraCore/cosmos-sdk/simapp"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/x/ibc/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc     *codec.Codec
	ctx     sdk.Context
	keeper  *keeper.Keeper
	querier sdk.Querier
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)
	legacyQuerierCdc := codec.NewAminoCodec(app.Codec())

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.keeper = app.IBCKeeper
	suite.querier = keeper.NewQuerier(*app.IBCKeeper, legacyQuerierCdc)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
