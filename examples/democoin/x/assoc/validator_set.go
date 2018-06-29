package assoc

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// ValidatorSet defines
type ValidatorSet struct {
	sdk.ValidatorSet

	store sdk.KVStore
	cdc   *wire.Codec

	maxAssoc int
	addrLen  int
}

var _ sdk.ValidatorSet = ValidatorSet{}

// NewValidatorSet returns new ValidatorSet with underlying ValidatorSet
func NewValidatorSet(cdc *wire.Codec, store sdk.KVStore, valset sdk.ValidatorSet, maxAssoc int, addrLen int) ValidatorSet {
	if maxAssoc < 0 || addrLen < 0 {
		panic("Cannot use negative integer for NewValidatorSet")
	}
	return ValidatorSet{
		ValidatorSet: valset,

		store: store,
		cdc:   cdc,

		maxAssoc: maxAssoc,
		addrLen:  addrLen,
	}
}

// Implements sdk.ValidatorSet
func (valset ValidatorSet) Validator(ctx sdk.Context, addr sdk.AccAddress) (res sdk.Validator) {
	base := valset.store.Get(GetBaseKey(addr))
	res = valset.ValidatorSet.Validator(ctx, base)
	if res == nil {
		res = valset.ValidatorSet.Validator(ctx, addr)
	}
	return
}

// GetBaseKey :: sdk.AccAddress -> sdk.AccAddress
func GetBaseKey(addr sdk.AccAddress) []byte {
	return append([]byte{0x00}, addr...)
}

// GetAssocPrefix :: sdk.AccAddress -> (sdk.AccAddress -> byte)
func GetAssocPrefix(base sdk.AccAddress) []byte {
	return append([]byte{0x01}, base...)
}

// GetAssocKey :: (sdk.AccAddress, sdk.AccAddress) -> byte
func GetAssocKey(base sdk.AccAddress, assoc sdk.AccAddress) []byte {
	return append(append([]byte{0x01}, base...), assoc...)
}

// Associate associates new address with validator address
func (valset ValidatorSet) Associate(ctx sdk.Context, base sdk.AccAddress, assoc sdk.AccAddress) bool {
	if len(base) != valset.addrLen || len(assoc) != valset.addrLen {
		return false
	}
	// If someone already owns the associated address
	if valset.store.Get(GetBaseKey(assoc)) != nil {
		return false
	}
	valset.store.Set(GetBaseKey(assoc), base)
	valset.store.Set(GetAssocKey(base, assoc), []byte{0x00})
	return true
}

// Dissociate removes association between addresses
func (valset ValidatorSet) Dissociate(ctx sdk.Context, base sdk.AccAddress, assoc sdk.AccAddress) bool {
	if len(base) != valset.addrLen || len(assoc) != valset.addrLen {
		return false
	}
	// No associated address found for given validator
	if !bytes.Equal(valset.store.Get(GetBaseKey(assoc)), base) {
		return false
	}
	valset.store.Delete(GetBaseKey(assoc))
	valset.store.Delete(GetAssocKey(base, assoc))
	return true
}

// Associations returns all associated addresses with a validator
func (valset ValidatorSet) Associations(ctx sdk.Context, base sdk.AccAddress) (res []sdk.AccAddress) {
	res = make([]sdk.AccAddress, valset.maxAssoc)
	iter := sdk.KVStorePrefixIterator(valset.store, GetAssocPrefix(base))
	i := 0
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		res[i] = key[len(key)-valset.addrLen:]
		i++
	}
	return res[:i]
}
