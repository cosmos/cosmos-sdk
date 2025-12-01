package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abnormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: math.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var govGenesis v1.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &govGenesis)

	const (
		minQuorum              = "0.200000000000000000"
		maxQuorum              = "0.760000000000000000"
		tallyThreshold         = "0.594000000000000000"
		amendmentMinQuorum     = "0.230000000000000000"
		amendmentMaxQuorum     = "0.820000000000000000"
		amendmentThreshold     = "0.947000000000000000"
		lawMinQuorum           = "0.200000000000000000"
		lawMaxQuorum           = "0.840000000000000000"
		lawThreshold           = "0.845000000000000000"
		burnDepositNoThreshold = "0.740000000000000000"
	)

	var (
		minDepositUpdatePeriod        = time.Duration(249813000000000)
		minInitialDepositUpdatePeriod = time.Duration(121469000000000)
	)

	require.Equal(t, []sdk.Coin{}, govGenesis.Params.MinDeposit)
	require.Equal(t, "52h44m19s", govGenesis.Params.MaxDepositPeriod.String())
	require.Equal(t, float64(278770), govGenesis.Params.VotingPeriod.Seconds())
	require.Equal(t, minQuorum, govGenesis.Params.QuorumRange.Min)
	require.Equal(t, maxQuorum, govGenesis.Params.QuorumRange.Max)
	require.Equal(t, tallyThreshold, govGenesis.Params.Threshold)
	require.Equal(t, amendmentMinQuorum, govGenesis.Params.ConstitutionAmendmentQuorumRange.Min)
	require.Equal(t, amendmentMaxQuorum, govGenesis.Params.ConstitutionAmendmentQuorumRange.Max)
	require.Equal(t, amendmentThreshold, govGenesis.Params.ConstitutionAmendmentThreshold)
	require.Equal(t, lawMinQuorum, govGenesis.Params.LawQuorumRange.Min)
	require.Equal(t, lawMaxQuorum, govGenesis.Params.LawQuorumRange.Max)
	require.Equal(t, lawThreshold, govGenesis.Params.LawThreshold)
	require.Equal(t, "", govGenesis.Params.MinInitialDepositRatio)
	require.Equal(t, "28h59m42s", govGenesis.Params.QuorumTimeout.String())
	require.Equal(t, "59h44m27s", govGenesis.Params.MaxVotingPeriodExtension.String())
	require.Equal(t, uint64(0xe), govGenesis.Params.QuorumCheckCount)
	require.Equal(t, burnDepositNoThreshold, govGenesis.Params.BurnDepositNoThreshold)
	require.Equal(t, uint64(0x28), govGenesis.StartingProposalId)
	require.Equal(t, []*v1.Deposit{}, govGenesis.Deposits)
	require.Equal(t, []*v1.Vote{}, govGenesis.Votes)
	require.Equal(t, []*v1.Proposal{}, govGenesis.Proposals)
	require.Equal(t, "", govGenesis.Constitution)
	require.Equal(t, v1.MinDepositThrottler{
		FloorValue:                        sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(631))),
		UpdatePeriod:                      &minDepositUpdatePeriod,
		TargetActiveProposals:             1,
		DecreaseSensitivityTargetDistance: 3,
		IncreaseRatio:                     "0.206000000000000000",
		DecreaseRatio:                     "0.050000000000000000",
	}, *govGenesis.Params.MinDepositThrottler)
	require.Equal(t, v1.MinInitialDepositThrottler{
		FloorValue:                        sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(201))),
		UpdatePeriod:                      &minInitialDepositUpdatePeriod,
		TargetProposals:                   29,
		DecreaseSensitivityTargetDistance: 1,
		IncreaseRatio:                     "0.247000000000000000",
		DecreaseRatio:                     "0.082000000000000000",
	}, *govGenesis.Params.MinInitialDepositThrottler)
}

// TestRandomizedGenState1 tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState1(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)
	// all these tests will panic
	tests := []struct {
		simState module.SimulationState
		panicMsg string
	}{
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       cdc,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}
