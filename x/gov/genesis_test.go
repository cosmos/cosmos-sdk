package gov_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

func TestImportExportQueues(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 2, valTokens)

	SortAddresses(addrs)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx = app.BaseApp.NewContext(false, abci.Header{})

	// Create two proposals, put the second into the voting period
	proposal := TestProposal
	proposal1, err := app.GovKeeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)
	proposalID1 := proposal1.ProposalID

	proposal2, err := app.GovKeeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)
	proposalID2 := proposal2.ProposalID

	votingStarted, err := app.GovKeeper.AddDeposit(ctx, proposalID2, addrs[0], app.GovKeeper.GetDepositParams(ctx).MinDeposit)
	require.NoError(t, err)
	require.True(t, votingStarted)

	proposal1, ok := app.GovKeeper.GetProposal(ctx, proposalID1)
	require.True(t, ok)
	proposal2, ok = app.GovKeeper.GetProposal(ctx, proposalID2)
	require.True(t, ok)
	require.True(t, proposal1.Status == gov.StatusDepositPeriod)
	require.True(t, proposal2.Status == gov.StatusVotingPeriod)

	authGenState := auth.ExportGenesis(ctx, app.AccountKeeper)
	bankGenState := bank.ExportGenesis(ctx, app.BankKeeper)

	// export the state and import it into a new app
	govGenState := gov.ExportGenesis(ctx, app.GovKeeper)
	genesisState := simapp.NewDefaultGenesisState()

	genesisState[auth.ModuleName] = app.AppCodec().MustMarshalJSON(authGenState)
	genesisState[bank.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenState)
	genesisState[gov.ModuleName] = app.AppCodec().MustMarshalJSON(govGenState)

	stateBytes, err := codec.MarshalJSONIndent(app.Codec(), genesisState)
	if err != nil {
		panic(err)
	}

	db := dbm.NewMemDB()
	app2 := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 0)

	app2.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	app2.Commit()
	app2.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app2.LastBlockHeight() + 1}})

	header = abci.Header{Height: app2.LastBlockHeight() + 1}
	app2.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx2 := app2.BaseApp.NewContext(false, abci.Header{})

	// Jump the time forward past the DepositPeriod and VotingPeriod
	ctx2 = ctx2.WithBlockTime(ctx2.BlockHeader().Time.Add(app2.GovKeeper.GetDepositParams(ctx2).MaxDepositPeriod).Add(app2.GovKeeper.GetVotingParams(ctx2).VotingPeriod))

	// Make sure that they are still in the DepositPeriod and VotingPeriod respectively
	proposal1, ok = app2.GovKeeper.GetProposal(ctx2, proposalID1)
	require.True(t, ok)
	proposal2, ok = app2.GovKeeper.GetProposal(ctx2, proposalID2)
	require.True(t, ok)
	require.True(t, proposal1.Status == gov.StatusDepositPeriod)
	require.True(t, proposal2.Status == gov.StatusVotingPeriod)

	macc := app2.GovKeeper.GetGovernanceAccount(ctx2)
	require.Equal(t, app2.GovKeeper.GetDepositParams(ctx2).MinDeposit, app2.BankKeeper.GetAllBalances(ctx2, macc.GetAddress()))

	// Run the endblocker. Check to make sure that proposal1 is removed from state, and proposal2 is finished VotingPeriod.
	gov.EndBlocker(ctx2, app2.GovKeeper)

	proposal1, ok = app2.GovKeeper.GetProposal(ctx2, proposalID1)
	require.False(t, ok)

	proposal2, ok = app2.GovKeeper.GetProposal(ctx2, proposalID2)
	require.True(t, ok)
	require.True(t, proposal2.Status == gov.StatusRejected)
}

func TestEqualProposals(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 2, valTokens)

	SortAddresses(addrs)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// Submit two proposals
	proposal := TestProposal
	proposal1, err := app.GovKeeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)

	proposal2, err := app.GovKeeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)

	// They are similar but their IDs should be different
	require.NotEqual(t, proposal1, proposal2)
	require.NotEqual(t, proposal1, proposal2)

	// Now create two genesis blocks
	state1 := gov.GenesisState{Proposals: []gov.Proposal{proposal1}}
	state2 := gov.GenesisState{Proposals: []gov.Proposal{proposal2}}
	require.NotEqual(t, state1, state2)
	require.False(t, state1.Equal(state2))

	// Now make proposals identical by setting both IDs to 55
	proposal1.ProposalID = 55
	proposal2.ProposalID = 55
	require.Equal(t, proposal1, proposal1)
	require.Equal(t, proposal1, proposal2)

	// Reassign proposals into state
	state1.Proposals[0] = proposal1
	state2.Proposals[0] = proposal2

	// State should be identical now..
	require.Equal(t, state1, state2)
	require.True(t, state1.Equal(state2))
}
