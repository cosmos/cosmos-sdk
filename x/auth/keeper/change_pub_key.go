package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	addressToPubKeyPrefix = []byte{0x9}
)

func (ak AccountKeeper) ChangePubKey(ctx sdk.Context, acc types.AccountI, pubKeyString string) error {
	pubKey, err := ak.StringToPubKey(pubKeyString)
	if err != nil {
		return fmt.Errorf("cannont convert pubKeyString to PubKey. Error: %w", err)
	}
	acc.SetPubKey(pubKey)

	store := ctx.KVStore(ak.storeKey)

	addressToPubKey := &types.AddressToPubKey{
		Address: acc.GetAddress().String(),
		PubKey:  pubKeyString,
	}

	bz, err := ak.cdc.Marshal(addressToPubKey)
	if err != nil {
		return fmt.Errorf("cannot marshal addressToPubKey, Error: %w", err)
	}

	store.Set(getAddressKey(acc.GetAddress().String()), bz)

	return nil
}

func (ak AccountKeeper) PubKeyFromStore(ctx sdk.Context, addr sdk.AccAddress) *types.AddressToPubKey {
	store := ctx.KVStore(ak.storeKey)

	bz := store.Get(getAddressKey(addr.String()))

	var addressToPubKey types.AddressToPubKey
	ak.cdc.MustUnmarshal(bz, &addressToPubKey)

	return &addressToPubKey
}

func getAddressKey(address string) []byte {
	return append(addressToPubKeyPrefix, []byte(address)...)
}

func (ak AccountKeeper) StringToPubKey(pubKeyString string) (cryptotypes.PubKey, error) {
	pubKeyBytes, _, err := crypto.UnarmorPubKeyBytes(pubKeyString)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	var pubKey cryptotypes.PubKey
	if err := ak.cdc.UnmarshalInterface(pubKeyBytes, &pubKey); err != nil {
		return nil, fmt.Errorf("cannont unmarshal public key: %w", err)
	}

	return pubKey, nil
}

func (ak AccountKeeper) GetAllPubKeys(ctx sdk.Context) []*types.AddressToPubKey {
	store := ctx.KVStore(ak.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, addressToPubKeyPrefix)
	defer iterator.Close()

	var pubkeyMappings []*types.AddressToPubKey
	for ; iterator.Valid(); iterator.Next() {
		var addressToPubKey types.AddressToPubKey
		ak.cdc.MustUnmarshal(iterator.Value(), &addressToPubKey)

		pubkeyMappings = append(pubkeyMappings, &addressToPubKey)
	}

	return pubkeyMappings
}

func (ak AccountKeeper) SavePubKeyMapping(ctx sdk.Context, pubKeyMapping *types.AddressToPubKey) {
	store := ctx.KVStore(ak.storeKey)

	key := getAddressKey(pubKeyMapping.Address)
	bz := ak.cdc.MustMarshal(pubKeyMapping)
	store.Set(key, bz)
}
