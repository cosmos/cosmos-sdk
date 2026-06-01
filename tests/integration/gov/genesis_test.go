package gov_test

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"gotest.tools/v3/assert"

	"cosmossdk.io/log/v2"

	sdkapp "github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client/flags"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestImportExportQueues(t *testing.T) {
	// app1: fully set up with default genesis
	app1 := testapp.Setup(t)
	cdc := app1.AppCodec()

	ctx1 := app1.NewContext(false)
	addrs := simtestutil.AddTestAddrs(app1.BankKeeper, app1.StakingKeeper, ctx1, 1, valTokens)

	_, err := app1.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app1.LastBlockHeight() + 1,
	})
	assert.NilError(t, err)

	ctx1 = app1.NewContext(false)

	// Create two proposals, put the second into the voting period
	proposal1, err := app1.GovKeeper.SubmitProposal(ctx1, []sdk.Msg{mkTestLegacyContent(t)}, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID1 := proposal1.Id

	proposal2, err := app1.GovKeeper.SubmitProposal(ctx1, []sdk.Msg{mkTestLegacyContent(t)}, "", "test", "description", addrs[0], false)
	assert.NilError(t, err)
	proposalID2 := proposal2.Id

	params, err := app1.GovKeeper.Params.Get(ctx1)
	assert.NilError(t, err)
	votingStarted, err := app1.GovKeeper.AddDeposit(ctx1, proposalID2, addrs[0], params.MinDeposit)
	assert.NilError(t, err)
	assert.Assert(t, votingStarted)

	proposal1, err = app1.GovKeeper.Proposals.Get(ctx1, proposalID1)
	assert.NilError(t, err)
	proposal2, err = app1.GovKeeper.Proposals.Get(ctx1, proposalID2)
	assert.NilError(t, err)
	assert.Assert(t, proposal1.Status == v1.StatusDepositPeriod)
	assert.Assert(t, proposal2.Status == v1.StatusVotingPeriod)

	// Export genesis from app1
	authGenState := app1.AccountKeeper.ExportGenesis(ctx1)
	bankGenState := app1.BankKeeper.ExportGenesis(ctx1)
	stakingGenState := app1.StakingKeeper.ExportGenesis(ctx1)
	distributionGenState := app1.DistrKeeper.ExportGenesis(ctx1)
	govGenState, _ := keeper.ExportGenesis(ctx1, &app1.GovKeeper)

	genesisState := app1.DefaultGenesis()
	genesisState[authtypes.ModuleName] = cdc.MustMarshalJSON(authGenState)
	genesisState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankGenState)
	genesisState[types.ModuleName] = cdc.MustMarshalJSON(govGenState)
	genesisState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingGenState)
	genesisState[disttypes.ModuleName] = cdc.MustMarshalJSON(distributionGenState)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	assert.NilError(t, err)

	// app2: fresh app initialized with exported genesis from app1
	opts2 := simtestutil.AppOptionsMap{
		flags.FlagHome:    t.TempDir(),
		flags.FlagChainID: "test-chain",
	}
	cfg2 := sdkapp.DefaultSDKAppConfig("app", opts2)
	app2 := sdkapp.NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg2)
	app2.LoadModules()
	err = app2.LoadLatestVersion()
	assert.NilError(t, err)

	_, err = app2.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
		ChainId:         "test-chain",
	})
	assert.NilError(t, err)

	_, err = app2.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app2.LastBlockHeight() + 1,
	})
	assert.NilError(t, err)

	_, err = app2.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app2.LastBlockHeight() + 1,
	})
	assert.NilError(t, err)

	ctx2 := app2.NewContext(false)

	params, err = app2.GovKeeper.Params.Get(ctx2)
	assert.NilError(t, err)
	// Jump the time forward past the DepositPeriod and VotingPeriod
	ctx2 = ctx2.WithBlockTime(ctx2.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod))

	// Make sure that they are still in the DepositPeriod and VotingPeriod respectively
	proposal1, err = app2.GovKeeper.Proposals.Get(ctx2, proposalID1)
	assert.NilError(t, err)
	proposal2, err = app2.GovKeeper.Proposals.Get(ctx2, proposalID2)
	assert.NilError(t, err)
	assert.Assert(t, proposal1.Status == v1.StatusDepositPeriod)
	assert.Assert(t, proposal2.Status == v1.StatusVotingPeriod)

	macc := app2.GovKeeper.GetGovernanceAccount(ctx2)
	assert.DeepEqual(t, sdk.Coins(params.MinDeposit), app2.BankKeeper.GetAllBalances(ctx2, macc.GetAddress()))

	// Run the endblocker. Check to make sure that proposal1 is removed from state, and proposal2 is finished VotingPeriod.
	err = gov.EndBlocker(ctx2, &app2.GovKeeper)
	assert.NilError(t, err)

	proposal1, err = app2.GovKeeper.Proposals.Get(ctx2, proposalID1)
	assert.ErrorContains(t, err, "not found")

	proposal2, err = app2.GovKeeper.Proposals.Get(ctx2, proposalID2)
	assert.NilError(t, err)
	assert.Assert(t, proposal2.Status == v1.StatusRejected)
}
