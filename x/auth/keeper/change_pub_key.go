package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	pubKeyMappingPrefix = []byte{0x9}
)

func (ak AccountKeeper) ChangePubKey(ctx sdk.Context, acc types.AccountI, pubKeyString string) error {
	now := time.Now()
	pubKey, err := ak.StringToPubKey(pubKeyString)
	if err != nil {
		return fmt.Errorf("cannont convert pubKeyString to PubKey. Error: %w", err)
	}
	acc.SetPubKey(pubKey)

	store := ctx.KVStore(ak.storeKey)

	oldPubKey := &types.PubKeyHistory{
		PubKey:    ak.PubKeyFromStore(ctx, acc.GetAddress()),
		ValidTill: &now,
	}

	pubKeyHistory := ak.GetPubKeyHistory(ctx, acc.GetAddress())

	pubKeyHistory = append(pubKeyHistory, oldPubKey)

	pubKeyMapping := &types.PubKeyMapping{
		Address:       acc.GetAddress().String(),
		PubKey:        pubKeyString,
		PubKeyHistory: pubKeyHistory,
	}

	bz, err := ak.cdc.Marshal(pubKeyMapping)
	if err != nil {
		return fmt.Errorf("cannot marshal pubKeyMapping, Error: %w", err)
	}

	amount := ak.GetParams(ctx).PubkeyChangeCost
	ctx.GasMeter().ConsumeGas(amount, "pubkey change fee")

	store.Set(getAddressKey(acc.GetAddress().String()), bz)

	return nil
}

func (ak AccountKeeper) GetPubKeyHistory(ctx sdk.Context, addr sdk.AccAddress) []*types.PubKeyHistory {
	store := ctx.KVStore(ak.storeKey)

	bz := store.Get(getAddressKey(addr.String()))

	var pubKeyMapping types.PubKeyMapping
	ak.cdc.MustUnmarshal(bz, &pubKeyMapping)

	return pubKeyMapping.PubKeyHistory
}

func (ak AccountKeeper) PubKeyFromStore(ctx sdk.Context, addr sdk.AccAddress) string {
	store := ctx.KVStore(ak.storeKey)

	bz := store.Get(getAddressKey(addr.String()))

	var pubKeyMapping types.PubKeyMapping
	ak.cdc.MustUnmarshal(bz, &pubKeyMapping)

	return pubKeyMapping.PubKey
}

func getAddressKey(address string) []byte {
	return append(pubKeyMappingPrefix, []byte(address)...)
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

func (ak AccountKeeper) GetAllPubKeys(ctx sdk.Context) []*types.PubKeyMapping {
	store := ctx.KVStore(ak.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, pubKeyMappingPrefix)
	defer iterator.Close()

	var pubkeyMappings []*types.PubKeyMapping
	for ; iterator.Valid(); iterator.Next() {
		var pubKeyMapping types.PubKeyMapping
		ak.cdc.MustUnmarshal(iterator.Value(), &pubKeyMapping)

		pubkeyMappings = append(pubkeyMappings, &pubKeyMapping)
	}

	return pubkeyMappings
}

func (ak AccountKeeper) SavePubKeyMapping(ctx sdk.Context, pubKeyMapping *types.PubKeyMapping) {
	store := ctx.KVStore(ak.storeKey)

	key := getAddressKey(pubKeyMapping.Address)
	bz := ak.cdc.MustMarshal(pubKeyMapping)
	store.Set(key, bz)
}
