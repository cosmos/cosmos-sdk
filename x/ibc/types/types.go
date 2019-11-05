package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the IBC module
	ModuleName = "ibc"

	// StoreKey is the string store representation
	StoreKey string = ModuleName

	// QuerierRoute is the querier route for the IBC module
	QuerierRoute string = ModuleName

	// RouterKey is the msg router key for the IBC module
	RouterKey string = ModuleName

	// DefaultCodespace of the IBC module
	DefaultCodespace sdk.CodespaceType = ModuleName
)
