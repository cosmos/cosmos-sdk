package gov_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestImportExportQueues(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 2, valTokens)

	SortAddresses(addrs)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	// Create two proposals, put the second into the voting period
	proposal1, err := app.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, nil)
	require.NoError(t, err)
	proposalID1 := proposal1.ProposalId

	proposal2, err := app.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, nil)
	require.NoError(t, err)
	proposalID2 := proposal2.ProposalId

	votingStarted, err := app.GovKeeper.AddDeposit(ctx, proposalID2, addrs[0], app.GovKeeper.GetDepositParams(ctx).MinDeposit)
	require.NoError(t, err)
	require.True(t, votingStarted)

	proposal1, ok := app.GovKeeper.GetProposal(ctx, proposalID1)
	require.True(t, ok)
	proposal2, ok = app.GovKeeper.GetProposal(ctx, proposalID2)
	require.True(t, ok)
	require.True(t, proposal1.Status == v1beta2.StatusDepositPeriod)
	require.True(t, proposal2.Status == v1beta2.StatusVotingPeriod)

	authGenState := auth.ExportGenesis(ctx, app.AccountKeeper)
	bankGenState := app.BankKeeper.ExportGenesis(ctx)
	stakingGenState := staking.ExportGenesis(ctx, app.StakingKeeper)
	distributionGenState := app.DistrKeeper.ExportGenesis(ctx)

	// export the state and import it into a new app
	govGenState := gov.ExportGenesis(ctx, app.GovKeeper)
	genesisState := simapp.NewDefaultGenesisState(app.AppCodec())

	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenState)
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenState)
	genesisState[types.ModuleName] = app.AppCodec().MustMarshalJSON(govGenState)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenState)
	genesisState[distributiontypes.ModuleName] = app.AppCodec().MustMarshalJSON(distributionGenState)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	db := dbm.NewMemDB()
	app2 := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 0, simapp.MakeTestEncodingConfig(), simapp.EmptyAppOptions{})

	app2.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simapp.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	app2.Commit()
	app2.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app2.LastBlockHeight() + 1}})

	header = tmproto.Header{Height: app2.LastBlockHeight() + 1}
	app2.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx2 := app2.BaseApp.NewContext(false, tmproto.Header{})

	// Jump the time forward past the DepositPeriod and VotingPeriod
	ctx2 = ctx2.WithBlockTime(ctx2.BlockHeader().Time.Add(*app2.GovKeeper.GetDepositParams(ctx2).MaxDepositPeriod).Add(*app2.GovKeeper.GetVotingParams(ctx2).VotingPeriod))

	// Make sure that they are still in the DepositPeriod and VotingPeriod respectively
	proposal1, ok = app2.GovKeeper.GetProposal(ctx2, proposalID1)
	require.True(t, ok)
	proposal2, ok = app2.GovKeeper.GetProposal(ctx2, proposalID2)
	require.True(t, ok)
	require.True(t, proposal1.Status == v1beta2.StatusDepositPeriod)
	require.True(t, proposal2.Status == v1beta2.StatusVotingPeriod)

	macc := app2.GovKeeper.GetGovernanceAccount(ctx2)
	require.Equal(t, sdk.Coins(app2.GovKeeper.GetDepositParams(ctx2).MinDeposit), app2.BankKeeper.GetAllBalances(ctx2, macc.GetAddress()))

	// Run the endblocker. Check to make sure that proposal1 is removed from state, and proposal2 is finished VotingPeriod.
	gov.EndBlocker(ctx2, app2.GovKeeper)

	proposal1, ok = app2.GovKeeper.GetProposal(ctx2, proposalID1)
	require.False(t, ok)

	proposal2, ok = app2.GovKeeper.GetProposal(ctx2, proposalID2)
	require.True(t, ok)
	require.True(t, proposal2.Status == v1beta2.StatusRejected)
}

func TestImportExportQueues_ErrorUnconsistentState(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	require.Panics(t, func() {
		gov.InitGenesis(ctx, app.AccountKeeper, app.BankKeeper, app.GovKeeper, &v1beta2.GenesisState{
			Deposits: v1beta2.Deposits{
				{
					ProposalId: 1234,
					Depositor:  "me",
					Amount: sdk.Coins{
						sdk.NewCoin(
							"stake",
							sdk.NewInt(1234),
						),
					},
				},
			},
		})
	})
}
