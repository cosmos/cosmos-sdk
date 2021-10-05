package v043

import (
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v040"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gov"
)

type (
	GenesisState = v040gov.GenesisState

	Proposals = v040gov.Proposals
)
