package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}
