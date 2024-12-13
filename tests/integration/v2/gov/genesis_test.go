package gov

import (
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/header"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	sdkmath "cosmossdk.io/math"
	_ "cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus"
	"cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	_ "cosmossdk.io/x/staking"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestImportExportQueues(t *testing.T) {
	var err error

	s1 := createTestSuite(t, integration.Genesis_COMMIT)
	ctx := s1.ctx

	addrs := simtestutil.AddTestAddrs(s1.BankKeeper, s1.StakingKeeper, ctx, 1, valTokens)

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

	authGenState, err := s1.AuthKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	bankGenState, err := s1.BankKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	stakingGenState, err := s1.StakingKeeper.ExportGenesis(ctx)
	require.NoError(t, err)

	// export the state and import it into a new app
	govGenState, err := gov.ExportGenesis(ctx, s1.GovKeeper)
	require.NoError(t, err)

	genesisState := s1.app.DefaultGenesis()

	genesisState[authtypes.ModuleName] = s1.cdc.MustMarshalJSON(authGenState)
	genesisState[banktypes.ModuleName] = s1.cdc.MustMarshalJSON(bankGenState)
	genesisState[types.ModuleName] = s1.cdc.MustMarshalJSON(govGenState)
	genesisState[stakingtypes.ModuleName] = s1.cdc.MustMarshalJSON(stakingGenState)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	assert.NilError(t, err)

	s2 := createTestSuite(t, integration.Genesis_SKIP)

	emptyHash := sha256.Sum256(nil)
	_, newstate, err := s2.app.InitGenesis(
		ctx,
		&server.BlockRequest[transaction.Tx]{
			Height:    1,
			Time:      time.Now(),
			Hash:      emptyHash[:],
			ChainId:   "test-chain",
			AppHash:   emptyHash[:],
			IsGenesis: true,
		},
		stateBytes,
		integration.NewGenesisTxCodec(s2.txConfigOptions),
	)
	assert.NilError(t, err)

	_, err = s2.app.Commit(newstate)
	assert.NilError(t, err)

	ctx2 := s2.app.StateLatestContext(t)

	params, err = s2.GovKeeper.Params.Get(ctx2)
	assert.NilError(t, err)
	// Jump the time forward past the DepositPeriod and VotingPeriod
	h := integration.HeaderInfoFromContext(ctx2)
	ctx2 = integration.SetHeaderInfo(ctx2, header.Info{Time: h.Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)})

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

func TestImportExportQueues_ErrorUnconsistentState(t *testing.T) {
	suite := createTestSuite(t, integration.Genesis_COMMIT)
	ctx := suite.ctx

	params := v1.DefaultParams()
	err := gov.InitGenesis(ctx, suite.AuthKeeper, suite.BankKeeper, suite.GovKeeper, &v1.GenesisState{
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
	err = gov.InitGenesis(ctx, suite.AuthKeeper, suite.BankKeeper, suite.GovKeeper, v1.DefaultGenesisState())
	require.NoError(t, err)
	genState, err := gov.ExportGenesis(ctx, suite.GovKeeper)
	require.NoError(t, err)
	require.Equal(t, genState, v1.DefaultGenesisState())
}
