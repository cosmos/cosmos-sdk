package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
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
	oneYear := now.AddDate(1, 0, 0)
	msgSrvr := keeper.NewMsgServerImpl(suite.keeper)

	allowance := &feegrant.BasicAllowance{SpendLimit: coins, Expiration: &oneYear}
	err := suite.keeper.GrantAllowance(suite.ctx, granterAddr, granteeAddr, allowance)
	suite.Require().NoError(err)

	genesis, err := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	// revoke fee allowance
	_, err = msgSrvr.RevokeAllowance(sdk.WrapSDKContext(suite.ctx), &feegrant.MsgRevokeAllowance{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	err = suite.keeper.InitGenesis(suite.ctx, genesis)
	suite.Require().NoError(err)

	newGenesis, err := suite.keeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(genesis, newGenesis)
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	any, err := codectypes.NewAnyWithValue(&testdata.Dog{})
	suite.Require().NoError(err)

	testCases := []struct {
		name          string
		feeAllowances []feegrant.Grant
	}{
		{
			"invalid granter",
			[]feegrant.Grant{
				{
					Granter: "invalid granter",
					Grantee: granteeAddr.String(),
				},
			},
		},
		{
			"invalid grantee",
			[]feegrant.Grant{
				{
					Granter: granterAddr.String(),
					Grantee: "invalid grantee",
				},
			},
		},
		{
			"invalid allowance",
			[]feegrant.Grant{
				{
					Granter:   granterAddr.String(),
					Grantee:   granteeAddr.String(),
					Allowance: any,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			err := suite.keeper.InitGenesis(suite.ctx, &feegrant.GenesisState{Allowances: tc.feeAllowances})
			suite.Require().Error(err)
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
