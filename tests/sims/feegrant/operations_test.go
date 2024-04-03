package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	_ "cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	_ "cosmossdk.io/x/auth/tx/config"
	_ "cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/keeper"
	_ "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/feegrant/simulation"
	_ "cosmossdk.io/x/mint"
	_ "cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/genutil"
)

type SimTestSuite struct {
	suite.Suite

	app               *runtime.App
	ctx               sdk.Context
	feegrantKeeper    keeper.Keeper
	interfaceRegistry codectypes.InterfaceRegistry
	txConfig          client.TxConfig
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	cdc               codec.Codec
	legacyAmino       *codec.LegacyAmino
}

func (suite *SimTestSuite) SetupTest() {
	var err error
	suite.app, err = simtestutil.Setup(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
				configurator.TxModule(),
				configurator.ConsensusModule(),
				configurator.GenutilModule(),
				configurator.FeegrantModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&suite.feegrantKeeper,
		&suite.bankKeeper,
		&suite.accountKeeper,
		&suite.interfaceRegistry,
		&suite.txConfig,
		&suite.cdc,
		&suite.legacyAmino,
	)
	suite.Require().NoError(err)

	suite.ctx = suite.app.BaseApp.NewContext(false).WithHeaderInfo(header.Info{Time: time.Now()})
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)
	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		err := banktestutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins)
		suite.Require().NoError(err)
	}

	return accounts
}

func (suite *SimTestSuite) TestWeightedOperations() {
	require := suite.Require()

	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(
		suite.interfaceRegistry,
		appParams, suite.cdc, suite.txConfig, suite.accountKeeper,
		suite.bankKeeper, suite.feegrantKeeper, codecaddress.NewBech32Codec("cosmos"),
	)

	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{
			simulation.DefaultWeightGrantAllowance,
			feegrant.ModuleName,
			sdk.MsgTypeURL(&feegrant.MsgGrantAllowance{}),
		},
		{
			simulation.DefaultWeightRevokeAllowance,
			feegrant.ModuleName,
			sdk.MsgTypeURL(&feegrant.MsgRevokeAllowance{}),
		},
	}

	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(r, suite.app.BaseApp, suite.ctx.WithHeaderInfo(header.Info{Time: time.Now()}), accs, suite.ctx.ChainID())
		require.NoError(err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

func (suite *SimTestSuite) TestSimulateMsgGrantAllowance() {
	app, ctx := suite.app, suite.ctx
	require := suite.Require()

	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)
	addr1, err := suite.accountKeeper.AddressCodec().BytesToString(accounts[1].Address)
	require.NoError(err)
	addr2, err := suite.accountKeeper.AddressCodec().BytesToString(accounts[2].Address)
	require.NoError(err)

	// execute operation
	op := simulation.SimulateMsgGrantAllowance(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.feegrantKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx.WithHeaderInfo(header.Info{Time: time.Now()}), accounts, "")
	require.NoError(err)

	var msg feegrant.MsgGrantAllowance
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal(addr2, msg.Granter)
	require.Equal(addr1, msg.Grantee)
	require.Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateMsgRevokeAllowance() {
	app, ctx := suite.app, suite.ctx
	require := suite.Require()

	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	feeAmt := sdk.TokensFromConsensusPower(200000, sdk.DefaultPowerReduction)
	feeCoins := sdk.NewCoins(sdk.NewCoin("foo", feeAmt))

	granter, grantee := accounts[0], accounts[1]

	oneYear := ctx.HeaderInfo().Time.AddDate(1, 0, 0)
	err := suite.feegrantKeeper.GrantAllowance(
		ctx,
		granter.Address,
		grantee.Address,
		&feegrant.BasicAllowance{
			SpendLimit: feeCoins,
			Expiration: &oneYear,
		},
	)
	require.NoError(err)

	granterStr, err := suite.accountKeeper.AddressCodec().BytesToString(accounts[0].Address)
	require.NoError(err)
	granteeStr, err := suite.accountKeeper.AddressCodec().BytesToString(accounts[1].Address)
	require.NoError(err)

	// execute operation
	op := simulation.SimulateMsgRevokeAllowance(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.feegrantKeeper, codecaddress.NewBech32Codec("cosmos"))
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(err)

	var msg feegrant.MsgRevokeAllowance
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal(granterStr, msg.Granter)
	require.Equal(granteeStr, msg.Grantee)
	require.Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
