package types

import (
	sdk "github.com/Stride-Labs/cosmos-sdk/types"
)

// invariant route
type InvarRoute struct {
	ModuleName string
	Route      string
	Invar      sdk.Invariant
}

// NewInvarRoute - create an InvarRoute object
func NewInvarRoute(moduleName, route string, invar sdk.Invariant) InvarRoute {
	return InvarRoute{
		ModuleName: moduleName,
		Route:      route,
		Invar:      invar,
	}
}

// get the full invariance route
func (i InvarRoute) FullRoute() string {
	return i.ModuleName + "/" + i.Route
}
