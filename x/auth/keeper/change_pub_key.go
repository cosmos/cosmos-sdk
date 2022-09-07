package keeper

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	addressToPubKeyPrefix = []byte{0x9}
)

func (ak AccountKeeper) ChangePubKey(ctx sdk.Context, acc types.AccountI, pubKey cryptotypes.PubKey) error {
	acc.SetPubKey(pubKey)

	store := ctx.KVStore(ak.storeKey)

	bz, err := ak.cdc.MarshalInterface(pubKey)
	if err != nil {
		return fmt.Errorf("cannot marshal pubkey, Error: %w", err)
	}

	store.Set(getAddressKey(acc.GetAddress().String()), bz)

	return nil
}

func (ak AccountKeeper) GetPubKeyFromStore(ctx sdk.Context, addr sdk.AccAddress) (cryptotypes.PubKey, error) {
	store := ctx.KVStore(ak.storeKey)

	bz := store.Get(getAddressKey(addr.String()))

	var pubkey cryptotypes.PubKey
	err := ak.cdc.UnmarshalInterface(bz, pubkey)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal pubkey, Error: %w", err)
	}

	return pubkey, nil
}

func getAddressKey(address string) []byte {
	return append(addressToPubKeyPrefix, []byte(address)...)
}
