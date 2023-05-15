package v2

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/crisis/exported"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

const (
	ModuleName = "nft"
)

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, legacySubspace exported.Subspace, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	iterator1 := storetypes.KVStorePrefixIterator(store, NFTKey)

	for ; iterator1.Valid(); iterator1.Next() {
		var nftData nft.NFT
		if err := cdc.Unmarshal(iterator1.Value(), &nftData); err != nil {
			return err
		}

		nftData.SendEnabled = true

		bz, _ := cdc.Marshal(&nftData)

		store.Set(iterator1.Key(), bz)
	}

	iterator2 := storetypes.KVStorePrefixIterator(store, NFTOfClassByOwnerKey)

	for ; iterator2.Valid(); iterator2.Next() {
		var nftData nft.NFT
		if err := cdc.Unmarshal(iterator2.Value(), &nftData); err != nil {
			return err
		}

		nftData.SendEnabled = true

		bz, _ := cdc.Marshal(&nftData)

		store.Set(iterator2.Key(), bz)
	}

	return nil
}
