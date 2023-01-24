package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

const (
	SanctionAddresses   = "sanction-addresses"
	SanctionTempEntries = "sanction-temp-entries"
	SanctionParams      = "sanction-params"
)

// RandomSanctionedAddresses randomly selects accounts to be sanctioned.
//
// Each account has a:
// * 20% chance of being sanctioned,
// * 80% chance of being ignored.
func RandomSanctionedAddresses(r *rand.Rand, accounts []simtypes.Account) []string {
	var rv []string
	for _, acct := range accounts {
		if r.Int63n(5) == 0 {
			rv = append(rv, acct.Address.String())
		}
	}
	return rv
}

// RandomTempEntries randomly selects accounts to be temporarily sanctioned/unsanctioned.
//
// Each account has a:
// * 10% chance of having a temp sanction,
// * 10% chance of having a temp unsanction,
// * 80% chance of being ignored.
func RandomTempEntries(r *rand.Rand, accounts []simtypes.Account) []*sanction.TemporaryEntry {
	var rv []*sanction.TemporaryEntry
	for _, acct := range accounts {
		switch r.Int63n(10) {
		case 0:
			rv = append(rv, &sanction.TemporaryEntry{
				Address:    acct.Address.String(),
				ProposalId: uint64(r.Int63n(1000) + 1_000_000_000),
				Status:     sanction.TEMP_STATUS_SANCTIONED,
			})
		case 1:
			rv = append(rv, &sanction.TemporaryEntry{
				Address:    acct.Address.String(),
				ProposalId: uint64(r.Int63n(1000) + 2_000_000_000),
				Status:     sanction.TEMP_STATUS_UNSANCTIONED,
			})
		}
	}
	return rv
}

// RandomParams generates randomized parameters for the sanction module.
//
// ImmediateSanctionMinDeposit and ImmediateUnsanctionMinDeposit are decided individually.
// Each has a:
// * 20% chance of being empty/zero.
// * 80% chance of being between 1 and 1000 (inclusive) of the default bond denom.
func RandomParams(r *rand.Rand) *sanction.Params {
	rv := &sanction.Params{}
	if r.Int63n(5) != 0 {
		rv.ImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, r.Int63n(1_000)+1))
	}
	if r.Int63n(5) != 0 {
		rv.ImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, r.Int63n(1_000)+1))
	}
	return rv
}

// RandomizedGenState creates a randomized sanction genesis state and adds it to the
// provided simState's GenState map.
func RandomizedGenState(simState *module.SimulationState) {
	genState := &sanction.GenesisState{}

	// SanctionedAddresses
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SanctionAddresses, &genState.SanctionedAddresses, simState.Rand,
		func(r *rand.Rand) {
			genState.SanctionedAddresses = RandomSanctionedAddresses(r, simtypes.RandomAccounts(r, len(simState.Accounts)))
		},
	)

	// TemporaryEntries
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SanctionTempEntries, &genState.TemporaryEntries, simState.Rand,
		func(r *rand.Rand) {
			genState.TemporaryEntries = RandomTempEntries(r, simtypes.RandomAccounts(r, len(simState.Accounts)))
		},
	)

	// Params
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SanctionParams, &genState.Params, simState.Rand,
		func(r *rand.Rand) { genState.Params = RandomParams(r) },
	)

	simState.GenState[sanction.ModuleName] = simState.Cdc.MustMarshalJSON(genState)
}
