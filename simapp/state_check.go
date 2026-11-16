package simapp

import (
	"github.com/cosmos/cosmos-sdk/testutil/statecheck"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CollectionsModules returns every module of the app whose keeper exposes a
// collections schema, with the store key the schema reads. Every other KV
// store of the app is listed in StateCheckExcludedStores.
func (app *SimApp) CollectionsModules() []statecheck.Module {
	return []statecheck.Module{
		{Name: authtypes.ModuleName, StoreKey: app.GetKey(authtypes.StoreKey), Schema: app.AccountKeeper.Schema},
		{Name: banktypes.ModuleName, StoreKey: app.GetKey(banktypes.StoreKey), Schema: app.BankKeeper.Schema},
		// x/distribution keeps the previous proposer, reward records, withdraw
		// addresses and slash events under raw keys outside its collections.
		{Name: distrtypes.ModuleName, StoreKey: app.GetKey(distrtypes.StoreKey), Schema: app.DistrKeeper.Schema, RawPrefixes: [][]byte{
			distrtypes.ProposerKey,
			distrtypes.ValidatorOutstandingRewardsPrefix,
			distrtypes.DelegatorWithdrawAddrPrefix,
			distrtypes.DelegatorStartingInfoPrefix,
			distrtypes.ValidatorHistoricalRewardsPrefix,
			distrtypes.ValidatorCurrentRewardsPrefix,
			distrtypes.ValidatorAccumulatedCommissionPrefix,
			distrtypes.ValidatorSlashEventPrefix,
		}},
		{Name: epochstypes.ModuleName, StoreKey: app.GetKey(epochstypes.StoreKey), Schema: app.EpochsKeeper.Schema},
		{Name: evidencetypes.ModuleName, StoreKey: app.GetKey(evidencetypes.StoreKey), Schema: app.EvidenceKeeper.Schema},
		{Name: feegrant.ModuleName, StoreKey: app.GetKey(feegrant.StoreKey), Schema: app.FeeGrantKeeper.Schema},
		{Name: govtypes.ModuleName, StoreKey: app.GetKey(govtypes.StoreKey), Schema: app.GovKeeper.Schema},
		{Name: minttypes.ModuleName, StoreKey: app.GetKey(minttypes.StoreKey), Schema: app.MintKeeper.Schema},
	}
}

// StateCheckExcludedStores maps each KV store that CheckStateDecodes does not
// read to the reason it is left out.
func StateCheckExcludedStores() map[string]string {
	return map[string]string{
		stakingtypes.StoreKey:        "keeper reads raw store keys, not collections",
		slashingtypes.StoreKey:       "keeper reads raw store keys, not collections",
		upgradetypes.StoreKey:        "keeper reads raw store keys, not collections",
		authzkeeper.StoreKey:         "keeper reads raw store keys, not collections",
		consensusparamtypes.StoreKey: "keeper exports no collections schema",
	}
}

// CheckStateDecodes decodes every key and value in the stores of
// CollectionsModules and returns a result per module and an error listing
// each entry that does not decode with the codecs the keepers read it with.
func (app *SimApp) CheckStateDecodes(ctx sdk.Context) ([]statecheck.Result, error) {
	return statecheck.Check(ctx, app.CollectionsModules())
}
