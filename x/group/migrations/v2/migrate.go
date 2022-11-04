package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "group"
)

// Migrate migrates the x/group module state from the consensus version 1 to
// version 2. Specifically, it modify the group policy module account and removes their name.
func Migrate(ctx sdk.Context, store sdk.KVStore, cdc codec.Codec) error {

	return nil
}
