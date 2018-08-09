package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// InitGenesis initializes the keeper's address to pubkey map.
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	addrPubkeyMap := make(map[[tmhash.Size]byte]crypto.PubKey, len(data.Validators))

	for _, validator := range data.Validators {
		keeper.addPubkey(validator.GetPubKey(), addrPubkeyMap, false)
	}

	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(addrPubkeyMap)
	store.Set(getAddrPubkeyMapKey(), bz)
	return
}
