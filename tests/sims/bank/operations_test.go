package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	_ "cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/auth"
	_ "cosmossdk.io/x/auth/tx/config"
	_ "cosmossdk.io/x/bank"
	"cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/testutil"
	"cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus"
	_ "cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

type SimTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	accountKeeper types.AccountKeeper
	bankKeeper    keeper.Keeper
	cdc           codec.Codec
	txConfig      client.TxConfig
	app           *runtime.App
}

func (suite *SimTestSuite) SetupTest() {
	var (
		appBuilder *runtime.AppBuilder
		err        error
	)
	suite.app, err = simtestutil.Setup(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AccountsModule(),
				configurator.AuthModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
				configurator.ConsensusModule(),
				configurator.TxModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		), &suite.accountKeeper, &suite.bankKeeper, &suite.cdc, &suite.txConfig, &appBuilder)

	suite.NoError(err)

	suite.ctx = suite.app.BaseApp.NewContext(false)
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(testutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins))
	}

	return accounts
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
