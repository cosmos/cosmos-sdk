package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

// TODO: Codespace will be removed
const DefaultCodespace = 65534

type Keeper struct {
	bk bank.Keeper
	ch ibc.Channel
}

func NewKeeper(key sdk.StoreKey, bk bank.Keeper, ibck ibc.Keeper) Keeper {
	return Keeper{
		bk: bk,
		// Prefixing for the future compatibility
		ch: ibck.Channel(sdk.NewPrefixStoreGetter(key, []byte{0x00})),
	}
}
