package oracle

import (
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the oracle store
type Keeper struct {
	key sdk.StoreKey
	cdc *codec.Codec

	valset sdk.ValidatorSet

	supermaj sdk.Dec
	timeout  int64
}

// NewKeeper constructs a new keeper
func NewKeeper(key sdk.StoreKey, cdc *codec.Codec, valset sdk.ValidatorSet, supermaj sdk.Dec, timeout int64) Keeper {
	if timeout < 0 {
		panic("Timeout should not be negative")
	}

	return Keeper{
		key: key,
		cdc: cdc,

		valset: valset,

		supermaj: supermaj,
		timeout:  timeout,
	}
}

// InfoStatus - current status of an Info
type InfoStatus int8

// Define InfoStatus
const (
	Pending = InfoStatus(iota)
	Processed
	Timeout
)

// Info for each payload
type Info struct {
	Power      sdk.Dec
	Hash       []byte
	LastSigned int64
	Status     InfoStatus
}

// EmptyInfo construct an empty Info
func EmptyInfo(ctx sdk.Context) Info {
	return Info{
		Power:      sdk.ZeroDec(),
		Hash:       ctx.BlockHeader().ValidatorsHash,
		LastSigned: ctx.BlockHeight(),
		Status:     Pending,
	}
}

// Info returns the information about a payload
func (keeper Keeper) Info(ctx sdk.Context, p Payload) (res Info) {
	store := ctx.KVStore(keeper.key)

	key := GetInfoKey(p, keeper.cdc)
	bz := store.Get(key)
	if bz == nil {
		return EmptyInfo(ctx)
	}
	keeper.cdc.MustUnmarshalBinary(bz, &res)

	return
}

func (keeper Keeper) setInfo(ctx sdk.Context, p Payload, info Info) {
	store := ctx.KVStore(keeper.key)

	key := GetInfoKey(p, keeper.cdc)
	bz := keeper.cdc.MustMarshalBinary(info)
	store.Set(key, bz)
}

func (keeper Keeper) sign(ctx sdk.Context, p Payload, signer sdk.AccAddress) {
	store := ctx.KVStore(keeper.key)

	key := GetSignKey(p, signer, keeper.cdc)
	store.Set(key, signer)
}

func (keeper Keeper) signed(ctx sdk.Context, p Payload, signer sdk.AccAddress) bool {
	store := ctx.KVStore(keeper.key)

	key := GetSignKey(p, signer, keeper.cdc)
	return store.Has(key)
}

func (keeper Keeper) clearSigns(ctx sdk.Context, p Payload) {
	store := ctx.KVStore(keeper.key)

	prefix := GetSignPrefix(p, keeper.cdc)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
	iter.Close()
}
