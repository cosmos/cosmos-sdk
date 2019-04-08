package bank

import sdk "github.com/cosmos/cosmos-sdk/types"

// name of this module
const ModuleName = "bank"

// app module for bank
type AppModule struct {
	keeper Keeper
}

// function name
func (a AppModule) Name() {
	return ModuleName
}

var _ AppModule = sdk.Module
