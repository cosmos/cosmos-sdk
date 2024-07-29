package simapp

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis file.
func (app *SimApp[T]) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (servertypes.ExportedApp, error) {
	panic("not implemented")
}
