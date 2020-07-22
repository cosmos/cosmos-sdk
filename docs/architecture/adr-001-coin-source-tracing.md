# ADR 001: Coin Source Tracing

## Changelog

- 2020-07-09: Initial Draft

## Status

Proposed

## Context

The specification for IBC cross-chain fungible token transfers
([ICS20](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer)), needs to
be aware of the origin any token denomination in order to relay a `Packet` which contains the sender
and recipient addressed in the
[`FungibleTokenPacketData`](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#data-structures).

The Packet relay works as follows (per
[specification](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay)):

> - When acting as the **source** zone, the bridge module escrows an existing local asset
>   denomination on the sending chain and mints vouchers on the receiving chain.
> - When acting as the **sink** zone, the bridge module burns local vouchers on the sending chains
>   and unescrows the local asset denomination on the receiving chain.

In this context, when the sender of a cross-chain transfer is **not** the source where the tokens
were originated, the protocol prefixes the denomination with the port and channel identifiers in the
following format:

```typescript
prefix + denom = {destPortN}/{destChannelN}/.../{destPort0}/{destChannel0}/denom
```

Example: transfering `100 uatom` from port `HubPort` and channel `HubChannel` on the Hub to
Ethermint's port `EthermintPort` and channel `EthermintChannel` results in `100
EthermintPort/EthermintChannel/uatom`, where `EthermintPort/EthermintChannel/uatom` is the new
denomination on the receiving chain.

In the case those tokens are transfered back to the Hub (i.e the **source** chain), the prefix is
trimmed and the token denomination updated to the original one.

### Problem

The problem of adding additional information to the coin denomination is twofold:

1. The ever increasing length if tokens are transfered to zones other than the source:

If a token is transfered `n` times via IBC to a sink chain, the token denom will contain `n` pairs
of prefixes, as shown on the format example above. This poses a problem because, while port
and channel identifiers have a maximim length of 64 each, the SDK `Coin` type only accepts
denoms up to 64 characters. Thus, a single cross-chain token, which again, is composed by the port and channels identifiers plus the base denomination, can exceed the length validation for the SDK `Coins`.

This can result in undesired behaviors such as tokens not being abled to be transferred to multiple
sink chains if the denomination exceeds the length or unexpected `panics` due to denomination
validation failing on the receiving chain.

2. The existence of special characters and uppercase letters on the denomination:

In the SDK every time a `Coin` is initialized through the constructor function `NewCoin`, a validation
of a coin's denom is performed according to a
[Regex](https://github.com/cosmos/cosmos-sdk/blob/a940214a4923a3bf9a9161cd14bd3072299cd0c9/types/coin.go#L583),
where only lowercase alphanumeric characters are accepted. While this is desirable for native denoms
to keep a clean UX, it presents a challenge for IBC as ports and channels might be randomly
generated with special carracters and uppercases as per the [ICS 024 - Host
Requirements](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements#paths-identifiers-separators)
specification.

## Decision

Instead of adding the identifiers on the coin denomination directly, the proposed solution hashes the denomination prefix in order to get a consistent lenght for all the cross-chain fungible tokens. The new format will be the following:

```golang
ibcDenom = "ibc/" + SHA256 hash of the trace identifiers prefix + "/" + coin denomination
```

In order to 

```golang
// GetDenom retreives the full identifiers trace from the store.
func (k Keeper) GetTrace(ctx Context, traceHash []byte) string {
  store := ctx.KVStore(k.storeKey)
  bz := store.Get(types.KeyTrace(traceHash))
  if len(bz) == 0 {
    return ""
  }
  return string(bz)
}

// HasTrace checks if a the key with the given trace hash exists on the store.
func (k Keeper) HasTrace(ctx Context, traceHash []byte)  bool {
  store := ctx.KVStore(k.storeKey)
  return store.Has(types.KeyTrace(traceHash))
}

// SetTrace sets a new {trace hash -> trace} pair to the store.
func (k Keeper) SetTrace(ctx Context, traceHash []byte, trace string) {
  store := ctx.KVStore(k.storeKey)
  store.Set(types.KeyTrace(traceHash), []byte(trace))
}
```

```golang
func (k Keeper) UpdateTrace(ctx Context, portID, channelID, denom string) string {
  // Get each component
  denomSplit := strings.Split(denom, "/")

  var (
    baseDenom string
    trace     string
    traceHash tmbytes.HexBytes
  )

  // return if denomination doesn't have separators
  if denomSplit[0] == denom {
    baseDenom = denom
    trace = portID + "/" + channelID +"/"
  } else {
    baseDenom = denomSplit[2]
    traceHash = tmbytes.HexBytes(denomSplit[1])
    // Get the value from the map trace hash -> denom identifiers prefix
    trace = k.GetTrace(ctx, prefixHash)
    // prefix the identifiers to create the new trace
    trace = portID + "/" + channelID +"/" + trace + "/"
  }
  
  traceHash = tmbytes.HexBytes(tmhash.Sum(trace))

  if !k.HasTrace(traceHash) {
    k.SetTrace(traceHash)
  }

  denom = "ibc/"+ traceHash.String() + baseDenom
  return denom
}
```

<!-- TODO: updates to ICS20 -->

### Positive

- Clearer separation of the origin tracing behaviour of the token (transfer prefix) from the original
  `Coin` denomination
- Consistent validation of `Coin` fields
- Cleaner `Coin` denominations for IBC

### Negative

- Store each set of tracing denomination identifiers on the `ibc-transfer` module store.
- Additional genesis fields.
- Slightly increases the gas usage on cross-chain transfers (1 read + 1 write).

### Neutral

- Slight difference with the ICS20 spec

## References

- [ICS 20 - Fungible token
  transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer)
