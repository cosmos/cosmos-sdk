package gov_test

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	_ "cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	authtypes "cosmossdk.io/x/auth/types"
	_ "cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	_ "cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
)

type suite struct {
	cdc           codec.Codec
	app           *runtime.App
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	GovKeeper     *keeper.Keeper
	StakingKeeper *stakingkeeper.Keeper
	appBuilder    *runtime.AppBuilder
}

var appConfig = configurator.NewAppConfig(
	configurator.AuthModule(),
	configurator.StakingModule(),
	configurator.BankModule(),
	configurator.GovModule(),
	configurator.MintModule(),
	configurator.ConsensusModule(),
	configurator.ProtocolPoolModule(),
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
		&s1.AccountKeeper, &s1.BankKeeper, &s1.GovKeeper, &s1.StakingKeeper, &s1.cdc, &s1.appBuilder,
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
	proposal1, err := s1.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	assert.NilError(t, err)
	proposalID1 := proposal1.Id

	proposal2, err := s1.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
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

	authGenState, err := s1.AccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	bankGenState, err := s1.BankKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	stakingGenState, err := s1.StakingKeeper.ExportGenesis(ctx)
	require.NoError(t, err)

	// export the state and import it into a new app
	govGenState, _ := gov.ExportGenesis(ctx, s1.GovKeeper)
	genesisState := s1.appBuilder.DefaultGenesis()

	genesisState[authtypes.ModuleName] = s1.cdc.MustMarshalJSON(authGenState)
	genesisState[banktypes.ModuleName] = s1.cdc.MustMarshalJSON(bankGenState)
	genesisState[types.ModuleName] = s1.cdc.MustMarshalJSON(govGenState)
	genesisState[stakingtypes.ModuleName] = s1.cdc.MustMarshalJSON(stakingGenState)

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
		&s2.AccountKeeper, &s2.BankKeeper, &s2.GovKeeper, &s2.StakingKeeper, &s2.cdc, &s2.appBuilder,
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
	ctx2 = ctx2.WithHeaderInfo(header.Info{Time: ctx2.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)})

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
	err = s2.GovKeeper.EndBlocker(ctx2)
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

func TestImportExportQueues_ErrorUnconsistentState(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.app
	ctx := app.BaseApp.NewContext(false)

	params := v1.DefaultParams()
	err := gov.InitGenesis(ctx, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper, &v1.GenesisState{
		Deposits: v1.Deposits{
			{
				ProposalId: 1234,
				Depositor:  "me",
				Amount: sdk.Coins{
					sdk.NewCoin(
						"stake",
						sdkmath.NewInt(1234),
					),
				},
			},
		},
		Params: &params,
	})
	require.Error(t, err)
	err = gov.InitGenesis(ctx, suite.AccountKeeper, suite.BankKeeper, suite.GovKeeper, v1.DefaultGenesisState())
	require.NoError(t, err)
	genState, err := gov.ExportGenesis(ctx, suite.GovKeeper)
	require.NoError(t, err)
	require.Equal(t, genState, v1.DefaultGenesisState())
}
