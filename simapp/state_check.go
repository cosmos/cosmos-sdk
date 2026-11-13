package simapp

import (
	"github.com/cosmos/cosmos-sdk/testutil/statecheck"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// CollectionsModules returns every module of the app whose keeper exposes a
// collections schema, with the store key the schema reads. Modules that do not
// build their store on collections, or keep the schema unexported, are not
// listed.
func (app *SimApp) CollectionsModules() []statecheck.Module {
	return []statecheck.Module{
		{Name: authtypes.ModuleName, StoreKey: app.GetKey(authtypes.StoreKey), Schema: app.AccountKeeper.Schema},
		{Name: banktypes.ModuleName, StoreKey: app.GetKey(banktypes.StoreKey), Schema: app.BankKeeper.Schema},
		{Name: distrtypes.ModuleName, StoreKey: app.GetKey(distrtypes.StoreKey), Schema: app.DistrKeeper.Schema},
		{Name: epochstypes.ModuleName, StoreKey: app.GetKey(epochstypes.StoreKey), Schema: app.EpochsKeeper.Schema},
		{Name: evidencetypes.ModuleName, StoreKey: app.GetKey(evidencetypes.StoreKey), Schema: app.EvidenceKeeper.Schema},
		{Name: feegrant.ModuleName, StoreKey: app.GetKey(feegrant.StoreKey), Schema: app.FeeGrantKeeper.Schema},
		{Name: govtypes.ModuleName, StoreKey: app.GetKey(govtypes.StoreKey), Schema: app.GovKeeper.Schema},
		{Name: minttypes.ModuleName, StoreKey: app.GetKey(minttypes.StoreKey), Schema: app.MintKeeper.Schema},
	}
}

// CheckStateDecodes decodes every key and value in the stores of
// CollectionsModules and returns an error listing each entry that does not
// decode with the codecs the keepers read it with.
func (app *SimApp) CheckStateDecodes(ctx sdk.Context) error {
	return statecheck.Check(ctx, app.CollectionsModules())
}
