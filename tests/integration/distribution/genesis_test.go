package distribution_test

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/stretchr/testify/suite"

	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/distribution/keeper"
	"cosmossdk.io/x/distribution/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

type ImportExportSuite struct {
	suite.Suite

	cdc                codec.Codec
	app                *runtime.App
	addrs              []sdk.AccAddress
	AccountKeeper      authkeeper.AccountKeeper
	BankKeeper         bankkeeper.Keeper
	DistributionKeeper keeper.Keeper
	StakingKeeper      *stakingkeeper.Keeper
	appBuilder         *runtime.AppBuilder
}

func TestDistributionImportExport(t *testing.T) {
	suite.Run(t, new(ImportExportSuite))
}

func (s *ImportExportSuite) SetupTest() {
	var err error
	valTokens := sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	s.app, err = simtestutil.SetupWithConfiguration(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		simtestutil.DefaultStartUpConfig(),
		&s.AccountKeeper, &s.BankKeeper, &s.DistributionKeeper, &s.StakingKeeper,
		&s.cdc, &s.appBuilder,
	)
	s.Require().NoError(err)

	ctx := s.app.BaseApp.NewContext(false)
	s.addrs = simtestutil.AddTestAddrs(s.BankKeeper, s.StakingKeeper, ctx, 1, valTokens)

	_, err = s.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: s.app.LastBlockHeight() + 1,
	})
	s.Require().NoError(err)
}

func (s *ImportExportSuite) TestHappyPath() {
	ctx := s.app.NewContext(true)
	// Imagine a situation where rewards were, e.g. 100 / 3 = 33, but the fee collector sent 100 to the distribution module.
	// There're 99 tokens in rewards, but 100 in the module; let's simulate a situation where there are 34 tokens left in the module,
	// and a single validator has 33 tokens of rewards.
	rewards := sdk.NewDecCoinsFromCoins(sdk.NewCoin("stake", sdkmath.NewInt(33)))

	// We'll pretend s.addrs[0] is the fee collector module.
	err := s.BankKeeper.SendCoinsFromAccountToModule(ctx, s.addrs[0], types.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(34))))
	s.Require().NoError(err)

	validators, err := s.StakingKeeper.GetAllValidators(ctx)
	s.Require().NoError(err)
	val := validators[0]

	err = s.DistributionKeeper.AllocateTokensToValidator(ctx, val, rewards)
	s.Require().NoError(err)

	_, err = s.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: s.app.LastBlockHeight() + 1,
	})
	s.Require().NoError(err)

	valBz, err := s.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
	s.Require().NoError(err)
	outstanding, err := s.DistributionKeeper.ValidatorOutstandingRewards.Get(ctx, valBz)
	s.Require().NoError(err)
	s.Require().Equal(rewards, outstanding.Rewards)

	genesisState, err := s.app.ModuleManager.ExportGenesis(ctx)
	s.Require().NoError(err)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	s.Require().NoError(err)

	db := coretesting.NewMemDB()
	conf2 := simtestutil.DefaultStartUpConfig()
	conf2.DB = db
	app2, err := simtestutil.SetupWithConfiguration(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		conf2,
	)
	s.Require().NoError(err)

	s.clearDB(db)
	err = app2.CommitMultiStore().LoadLatestVersion()
	s.Require().NoError(err)

	_, err = app2.InitChain(
		&abci.InitChainRequest{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	s.Require().NoError(err)
}

func (s *ImportExportSuite) TestInsufficientFunds() {
	ctx := s.app.NewContext(true)
	rewards := sdk.NewCoin("stake", sdkmath.NewInt(35))

	validators, err := s.StakingKeeper.GetAllValidators(ctx)
	s.Require().NoError(err)

	err = s.DistributionKeeper.AllocateTokensToValidator(ctx, validators[0], sdk.NewDecCoinsFromCoins(rewards))
	s.Require().NoError(err)

	// We'll pretend s.addrs[0] is the fee collector module.
	err = s.BankKeeper.SendCoinsFromAccountToModule(ctx, s.addrs[0], types.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(34))))
	s.Require().NoError(err)

	_, err = s.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: s.app.LastBlockHeight() + 1,
	})
	s.Require().NoError(err)

	genesisState, err := s.app.ModuleManager.ExportGenesis(ctx)
	s.Require().NoError(err)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	s.Require().NoError(err)

	db := coretesting.NewMemDB()
	conf2 := simtestutil.DefaultStartUpConfig()
	conf2.DB = db
	app2, err := simtestutil.SetupWithConfiguration(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		conf2,
	)
	s.Require().NoError(err)

	s.clearDB(db)
	err = app2.CommitMultiStore().LoadLatestVersion()
	s.Require().NoError(err)

	_, err = app2.InitChain(
		&abci.InitChainRequest{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	s.Require().ErrorContains(err, "distribution module balance is less than module holdings")
}

func (s *ImportExportSuite) clearDB(db corestore.KVStoreWithBatch) {
	iter, err := db.Iterator(nil, nil)
	s.Require().NoError(err)
	defer iter.Close()

	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}

	for _, k := range keys {
		s.Require().NoError(db.Delete(k))
	}
}
