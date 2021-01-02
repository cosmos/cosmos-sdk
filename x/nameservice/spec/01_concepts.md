<!--
order: 1
-->

# Concepts


## Keeper

In each module, the Keeper manages state, specifies how an application’s state changes, and calls keepers from other modules if passed.

The Keepers in the `nameservice` module are:
- coinKeeper bank.keeper - Bank Module
- nameStoreKey sdk.StoreKey - Key to access our KVStore from Context
- cdc *codec.Codec - codec for binary encoding/decoding

```protobuf
// Keeper of the nameservice store
type Keeper struct {
	CoinKeeper bank.Keeper
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	// paramspace types.ParamSubspace
}
```
## Keeper Actions

Get - resolves the name, checks if the name has a owner, gets the owner, and the finally getts the price.

Set - sets the name if it is available, sets the owner, and sets the price.

Delete - lets the owner of the name to delete it.

**Example**:

```protobuf
// GetWhois returns the whois information
func (k Keeper) GetWhois(ctx sdk.Context, key string) (types.Whois, error) {
	store := ctx.KVStore(k.storeKey)
	var whois types.Whois
	byteKey := []byte(types.WhoisPrefix + key)
	err := k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(byteKey), &whois)
	if err != nil {
		return whois, err
	}
	return whois, nil
}

// SetWhois sets a whois. We modified this function to use the `name` value as the key instead of msg.ID
func (k Keeper) SetWhois(ctx sdk.Context, name string, whois types.Whois) {

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(whois)
	key := []byte(types.WhoisPrefix + name)
	store.Set(key, bz)
}
// DeleteWhois deletes a whois
func (k Keeper) DeleteWhois(ctx sdk.Context, key string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete([]byte(types.WhoisPrefix + key))
}
```