package gov_test

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type suite struct {
	cdc           codec.Codec
	app           *runtime.App
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	DistrKeeper   distrkeeper.Keeper
	GovKeeper     *keeper.Keeper
	StakingKeeper *stakingkeeper.Keeper
	appBuilder    *runtime.AppBuilder
}

var appConfig = configurator.NewAppConfig(
	configurator.ParamsModule(),
	configurator.AuthModule(),
	configurator.StakingModule(),
	configurator.BankModule(),
	configurator.GovModule(),
	configurator.DistributionModule(),
	configurator.MintModule(),
	configurator.ConsensusModule(),
)

func TestImportExportQueues(t *testing.T) {
	var err error

	s1 := suite{}
	s1.app, err = simtestutil.SetupWithConfiguration(
		depinject.Configs(
			appConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		simtestutil.DefaultStartUpConfig(),
		&s1.AccountKeeper, &s1.BankKeeper, &s1.DistrKeeper, &s1.GovKeeper, &s1.StakingKeeper, &s1.cdc, &s1.appBuilder,
	)
	assert.NilError(t, err)

	ctx := s1.app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(s1.BankKeeper, s1.StakingKeeper, ctx, 1, valTokens)

	_, err = s1.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s1.app.LastBlockHeight() + 1,
	})
	assert.NilError(t, err)

	ctx = s1.app.BaseApp.NewContext(false)
	// Create two proposals, put the second into the voting period
	proposal1, err := s1.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID1 := proposal1.Id

	proposal2, err := s1.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID2 := proposal2.Id

	params, err := s1.GovKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	votingStarted, err := s1.GovKeeper.AddDeposit(ctx, proposalID2, addrs[0], params.MinDeposit)
	assert.NilError(t, err)
	assert.Assert(t, votingStarted)

	proposal1, err = s1.GovKeeper.Proposals.Get(ctx, proposalID1)
	assert.NilError(t, err)
	proposal2, err = s1.GovKeeper.Proposals.Get(ctx, proposalID2)
	assert.NilError(t, err)
	assert.Assert(t, proposal1.Status == v1.StatusDepositPeriod)
	assert.Assert(t, proposal2.Status == v1.StatusVotingPeriod)

	authGenState := s1.AccountKeeper.ExportGenesis(ctx)
	bankGenState := s1.BankKeeper.ExportGenesis(ctx)
	stakingGenState := s1.StakingKeeper.ExportGenesis(ctx)
	distributionGenState := s1.DistrKeeper.ExportGenesis(ctx)

	// export the state and import it into a new app
	govGenState, _ := gov.ExportGenesis(ctx, s1.GovKeeper)
	genesisState := s1.appBuilder.DefaultGenesis()

	genesisState[authtypes.ModuleName] = s1.cdc.MustMarshalJSON(authGenState)
	genesisState[banktypes.ModuleName] = s1.cdc.MustMarshalJSON(bankGenState)
	genesisState[types.ModuleName] = s1.cdc.MustMarshalJSON(govGenState)
	genesisState[stakingtypes.ModuleName] = s1.cdc.MustMarshalJSON(stakingGenState)
	genesisState[disttypes.ModuleName] = s1.cdc.MustMarshalJSON(distributionGenState)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	assert.NilError(t, err)

	s2 := suite{}
	db := dbm.NewMemDB()
	conf2 := simtestutil.DefaultStartUpConfig()
	conf2.DB = db
	s2.app, err = simtestutil.SetupWithConfiguration(
		depinject.Configs(
			appConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		conf2,
		&s2.AccountKeeper, &s2.BankKeeper, &s2.DistrKeeper, &s2.GovKeeper, &s2.StakingKeeper, &s2.cdc, &s2.appBuilder,
	)
	assert.NilError(t, err)

	clearDB(t, db)
	err = s2.app.CommitMultiStore().LoadLatestVersion()
	assert.NilError(t, err)

	_, err = s2.app.InitChain(
		&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	assert.NilError(t, err)

	_, err = s2.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s2.app.LastBlockHeight() + 1,
	})
	assert.NilError(t, err)

	_, err = s2.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s2.app.LastBlockHeight() + 1,
	})
	assert.NilError(t, err)

	ctx2 := s2.app.BaseApp.NewContext(false)

	params, err = s2.GovKeeper.Params.Get(ctx2)
	assert.NilError(t, err)
	// Jump the time forward past the DepositPeriod and VotingPeriod
	ctx2 = ctx2.WithBlockTime(ctx2.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod))

	// Make sure that they are still in the DepositPeriod and VotingPeriod respectively
	proposal1, err = s2.GovKeeper.Proposals.Get(ctx2, proposalID1)
	assert.NilError(t, err)
	proposal2, err = s2.GovKeeper.Proposals.Get(ctx2, proposalID2)
	assert.NilError(t, err)
	assert.Assert(t, proposal1.Status == v1.StatusDepositPeriod)
	assert.Assert(t, proposal2.Status == v1.StatusVotingPeriod)

	macc := s2.GovKeeper.GetGovernanceAccount(ctx2)
	assert.DeepEqual(t, sdk.Coins(params.MinDeposit), s2.BankKeeper.GetAllBalances(ctx2, macc.GetAddress()))

	// Run the endblocker. Check to make sure that proposal1 is removed from state, and proposal2 is finished VotingPeriod.
	err = gov.EndBlocker(ctx2, s2.GovKeeper)
	assert.NilError(t, err)

	proposal1, err = s2.GovKeeper.Proposals.Get(ctx2, proposalID1)
	assert.ErrorContains(t, err, "not found")

	proposal2, err = s2.GovKeeper.Proposals.Get(ctx2, proposalID2)
	assert.NilError(t, err)
	assert.Assert(t, proposal2.Status == v1.StatusRejected)
}

func clearDB(t *testing.T, db *dbm.MemDB) {
	t.Helper()
	iter, err := db.Iterator(nil, nil)
	assert.NilError(t, err)
	defer iter.Close()

	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}

	for _, k := range keys {
		assert.NilError(t, db.Delete(k))
	}
}
