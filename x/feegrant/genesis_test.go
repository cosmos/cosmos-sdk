package feegrant_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	feegrant "github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

type GenesisTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
}

func (suite *GenesisTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1})
	suite.keeper = app.FeeGrantKeeper
}

var (
	granteePub  = secp256k1.GenPrivKey().PubKey()
	granterPub  = secp256k1.GenPrivKey().PubKey()
	granteeAddr = sdk.AccAddress(granteePub.Address())
	granterAddr = sdk.AccAddress(granterPub.Address())
)

func (suite *GenesisTestSuite) TestImportExportGenesis() {
	coins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1_000)))
	now := suite.ctx.BlockHeader().Time
	msgSrvr := keeper.NewMsgServerImpl(suite.keeper)

	allowance := &types.BasicFeeAllowance{SpendLimit: coins, Expiration: types.ExpiresAtTime(now.AddDate(1, 0, 0))}
	err := suite.keeper.GrantFeeAllowance(suite.ctx, granterAddr, granteeAddr, allowance)
	suite.Require().NoError(err)

	genesis, err := feegrant.ExportGenesis(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	// revoke fee allowance
	_, err = msgSrvr.RevokeFeeAllowance(sdk.WrapSDKContext(suite.ctx), &types.MsgRevokeFeeAllowance{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	feegrant.InitGenesis(suite.ctx, suite.keeper, genesis)
	newGenesis, err := feegrant.ExportGenesis(suite.ctx, suite.keeper)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis, newGenesis)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
