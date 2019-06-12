package v0_36

import (
	"github.com/cosmos/cosmos-sdk/codec"
	extypes "github.com/cosmos/cosmos-sdk/contrib/export/types"
	v034gov "github.com/cosmos/cosmos-sdk/contrib/export/v0_34/gov"
	"github.com/cosmos/cosmos-sdk/contrib/export/v0_36/gov"
)

func migrateGovernance(initialState v034gov.GenesisState) gov.GenesisState {
	var targetGov gov.GenesisState
	//TODO
	return targetGov
}

func Migrate(appState extypes.AppMap, cdc *codec.Codec) extypes.AppMap {
	var governance v034gov.GenesisState

	cdc.MustUnmarshalJSON(appState[`gov`], governance)
	appState["gov"] = cdc.MustMarshalJSON(migrateGovernance(governance))

	return appState
}
