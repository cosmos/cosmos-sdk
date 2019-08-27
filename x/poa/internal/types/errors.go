package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodeSpace sdk.CodespaceType = ModuleName
	// Using the errors from the staking module here
	// see file github.com/cosmos/cosmos-sdk/x/staking/types
)
