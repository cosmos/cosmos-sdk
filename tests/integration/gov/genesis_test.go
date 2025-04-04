package gov_test

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

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
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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
		appConfig,
		simtestutil.DefaultStartUpConfig(),
		&s1.AccountKeeper, &s1.BankKeeper, &s1.DistrKeeper, &s1.GovKeeper, &s1.StakingKeeper, &s1.cdc, &s1.appBuilder,
	)
	require.NoError(t, err)

	ctx := s1.app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simtestutil.AddTestAddrs(s1.BankKeeper, s1.StakingKeeper, ctx, 1, valTokens)

	header := tmproto.Header{Height: s1.app.LastBlockHeight() + 1}
	s1.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx = s1.app.BaseApp.NewContext(false, tmproto.Header{})
	// Create two proposals, put the second into the voting period
	proposal1, err := s1.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "test", "description", addrs[0])
	require.NoError(t, err)
	proposalID1 := proposal1.Id

	proposal2, err := s1.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "test", "description", addrs[0])
	require.NoError(t, err)
	proposalID2 := proposal2.Id

	votingStarted, err := s1.GovKeeper.AddDeposit(ctx, proposalID2, addrs[0], s1.GovKeeper.GetParams(ctx).MinDeposit)
	require.NoError(t, err)
	require.True(t, votingStarted)

	proposal1, ok := s1.GovKeeper.GetProposal(ctx, proposalID1)
	require.True(t, ok)
	proposal2, ok = s1.GovKeeper.GetProposal(ctx, proposalID2)
	require.True(t, ok)
	require.True(t, proposal1.Status == v1.StatusDepositPeriod)
	require.True(t, proposal2.Status == v1.StatusVotingPeriod)

	authGenState := s1.AccountKeeper.ExportGenesis(ctx)
	bankGenState := s1.BankKeeper.ExportGenesis(ctx)
	stakingGenState := s1.StakingKeeper.ExportGenesis(ctx)
	distributionGenState := s1.DistrKeeper.ExportGenesis(ctx)

	// export the state and import it into a new app
	govGenState := gov.ExportGenesis(ctx, s1.GovKeeper)
	genesisState := s1.appBuilder.DefaultGenesis()

	genesisState[authtypes.ModuleName] = s1.cdc.MustMarshalJSON(authGenState)
	genesisState[banktypes.ModuleName] = s1.cdc.MustMarshalJSON(bankGenState)
	genesisState[types.ModuleName] = s1.cdc.MustMarshalJSON(govGenState)
	genesisState[stakingtypes.ModuleName] = s1.cdc.MustMarshalJSON(stakingGenState)
	genesisState[distributiontypes.ModuleName] = s1.cdc.MustMarshalJSON(distributionGenState)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	s2 := suite{}
	s2.app, err = simtestutil.SetupWithConfiguration(
		appConfig,
		simtestutil.DefaultStartUpConfig(),
		&s2.AccountKeeper, &s2.BankKeeper, &s2.DistrKeeper, &s2.GovKeeper, &s2.StakingKeeper, &s2.cdc, &s2.appBuilder,
	)
	require.NoError(t, err)

	s2.app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	s2.app.Commit()
	s2.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s2.app.LastBlockHeight() + 1}})

	header = tmproto.Header{Height: s2.app.LastBlockHeight() + 1}
	s2.app.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx2 := s2.app.BaseApp.NewContext(false, tmproto.Header{})

	// Jump the time forward past the DepositPeriod and VotingPeriod
	ctx2 = ctx2.WithBlockTime(ctx2.BlockHeader().Time.Add(*s2.GovKeeper.GetParams(ctx2).MaxDepositPeriod).Add(*s2.GovKeeper.GetParams(ctx2).VotingPeriod))

	// Make sure that they are still in the DepositPeriod and VotingPeriod respectively
	proposal1, ok = s2.GovKeeper.GetProposal(ctx2, proposalID1)
	require.True(t, ok)
	proposal2, ok = s2.GovKeeper.GetProposal(ctx2, proposalID2)
	require.True(t, ok)
	require.True(t, proposal1.Status == v1.StatusDepositPeriod)
	require.True(t, proposal2.Status == v1.StatusVotingPeriod)

	macc := s2.GovKeeper.GetGovernanceAccount(ctx2)
	require.Equal(t, sdk.Coins(s2.GovKeeper.GetParams(ctx2).MinDeposit), s2.BankKeeper.GetAllBalances(ctx2, macc.GetAddress()))

	// Run the endblocker. Check to make sure that proposal1 is removed from state, and proposal2 is finished VotingPeriod.
	gov.EndBlocker(ctx2, s2.GovKeeper)

	proposal1, ok = s2.GovKeeper.GetProposal(ctx2, proposalID1)
	require.False(t, ok)

	proposal2, ok = s2.GovKeeper.GetProposal(ctx2, proposalID2)
	require.True(t, ok)
	require.True(t, proposal2.Status == v1.StatusRejected)
}
