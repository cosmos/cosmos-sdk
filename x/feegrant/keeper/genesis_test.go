package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/testutil"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

type GenesisTestSuite struct {
	suite.Suite
	ctx            sdk.Context
	feegrantKeeper keeper.Keeper
}

func (suite *GenesisTestSuite) SetupTest() {
	checkTx := false

	var (
		interfaceRegistry codectypes.InterfaceRegistry
		bankKeeper        bankkeeper.Keeper
		stakingKeeper     *stakingkeeper.Keeper
		cdc               codec.Codec
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&suite.feegrantKeeper,
		&bankKeeper,
		&stakingKeeper,
		&interfaceRegistry,
		&cdc,
	)
	suite.Require().NoError(err)
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1})
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
	msgSrvr := keeper.NewMsgServerImpl(suite.feegrantKeeper)

	allowance := &feegrant.BasicAllowance{SpendLimit: coins, Expiration: &oneYear}
	err := suite.feegrantKeeper.GrantAllowance(suite.ctx, granterAddr, granteeAddr, allowance)
	suite.Require().NoError(err)

	genesis, err := suite.feegrantKeeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	// revoke fee allowance
	_, err = msgSrvr.RevokeAllowance(sdk.WrapSDKContext(suite.ctx), &feegrant.MsgRevokeAllowance{
		Granter: granterAddr.String(),
		Grantee: granteeAddr.String(),
	})
	suite.Require().NoError(err)
	err = suite.feegrantKeeper.InitGenesis(suite.ctx, genesis)
	suite.Require().NoError(err)

	newGenesis, err := suite.feegrantKeeper.ExportGenesis(suite.ctx)
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
			err := suite.feegrantKeeper.InitGenesis(suite.ctx, &feegrant.GenesisState{Allowances: tc.feeAllowances})
			suite.Require().Error(err)
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
