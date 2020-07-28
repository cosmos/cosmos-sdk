# ADR 001: Coin Source Tracing

## Changelog

- 2020-07-09: Initial Draft

## Status

Proposed

## Context

The specification for IBC cross-chain fungible token transfers
([ICS20](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer)), needs to
be aware of the origin of any token denomination in order to relay a `Packet` which contains the sender
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

If a token is transferred `n` times via IBC to a sink chain, the token denom will contain `n` pairs
of prefixes, as shown on the format example above. This poses a problem because, while port and
channel identifiers have a maximum length of 64 each, the SDK `Coin` type only accepts denoms up to
64 characters. Thus, a single cross-chain token, which again, is composed by the port and channels
identifiers plus the base denomination, can exceed the length validation for the SDK `Coins`.

This can result in undesired behaviours such as tokens not being able to be transferred to multiple
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

Instead of adding the identifiers on the coin denomination directly, the proposed solution hashes
the denomination prefix in order to get a consistent length for all the cross-chain fungible tokens.
The new format will be the following:

```golang
ibcDenom = "ibc/" + hash(trace + "/" + base denom)
```

The hash function will be a SHA256 hash of the fields of the `DenomTrace`:

```protobuf
// DenomTrace contains the base denomination for ICS20 fungible tokens and the souce tracing
// information
message DenomTrace {
  // chain of port/channel identifiers used for tracing the souce of the fungible token
  string trace = 1;
  // base denomination of the relayed fungible token
  string base_denom = 2;
}
```

```golang
// Hash returns the hex bytes of the SHA256 hash of the DenomTrace fields.
func (dt DenomTrace) Hash() tmbytes.HexBytes {
  return tmhash.Sum(dt.Trace + "/" + dt.BaseDenom)
}
```

### `x/ibc-transfer` Changes

In order to retreive the trace information from an IBC denomination, a lookup table needs to be
added to the `ibc-transfer` module. These values need to also be persisted between upgrades, meaning
that a new `[]DenomTrace` `GenesisState` field state needs to be added to the module:

```golang
// GetDenom retreives the full identifiers trace and base denomination from the store.
func (k Keeper) GetDenomTrace(ctx Context, denomTraceHash []byte) (DenomTrace, bool) {
  store := ctx.KVStore(k.storeKey)
  bz := store.Get(types.KeyDenomTrace(traceHash))
  if bz == nil {
    return &DenomTrace, false
  }

  var denomTrace DenomTrace
  k.cdc.MustUnmarshalBinaryBare(bz, &denomTrace)
  return denomTrace, true
}

// HasTrace checks if a the key with the given trace hash exists on the store.
func (k Keeper) HasDenomTrace(ctx Context, denomTraceHash []byte)  bool {
  store := ctx.KVStore(k.storeKey)
  return store.Has(types.KeyTrace(denomTraceHash))
}

// SetTrace sets a new {trace hash -> trace} pair to the store.
func (k Keeper) SetTrace(ctx Context, denomTraceHash []byte, denomTrace DenomTrace) {
  store := ctx.KVStore(k.storeKey)
  bz := k.cdc.MustMarshalBinaryBare(&denomTrace)
  store.Set(types.KeyTrace(traceHash), bz)
}
```

When a fungible token is sent to a sink chain, the trace information needs to be updated with the
new port and channel identifiers:

```golang
// PrefixDenom adds the given port and channel identifiers prefix to the denomination and sets the
// new {trace hash -> trace} pair to the store.
func (k Keeper) PrefixDenom(ctx Context, portID, channelID, denom string) string {
  // Get each component of the denom. The resulting slice will be:
  //
  // - ["ibc", hash], if the denom is dirty (contains trace metadata).
  // - [baseDenom], if the denom has never been sent from the origin chain.
  denomSplit := strings.SplitN(denom, "/", 3)

  var (
    baseDenom string
    trace     string
    traceHash tmbytes.HexBytes
  )

  // check if the denomination is clean or if it contains the trace info
  if len(denomSplit) == 1 && denomSplit[0] == denom {
    baseDenom = denom
    trace = portID + "/" + channelID +"/"
  } else {
    traceHash = tmbytes.HexBytes(denomSplit[1])
    // Get the value from the map trace hash -> trace info prefix
    denomTrace, found := k.GetDenomTrace(ctx, traceHash)
    if found {
      // prefix the identifiers to create the new trace
      trace = portID + "/" + channelID +"/" + denomTrace.Trace + "/"
      baseDenom = denomTrace.BaseDenom
    } else {
      // TODO: construct the trace info from the msg fields
    }
  }

  denomTrace = DenomTrace{Trace: trace, BaseDenom: baseDenom}
  
  traceHash = denomTrace.Hash()

  // set the value to the lookup table if not stored already
  if !k.HasDenomTrace(ctx, traceHash) {
    k.SetDenomTrace(ctx, traceHash, denomTrace)
  }

  denom = "ibc/"+ traceHash.String()
  return denom
}
```

The denomination also needs to be updated when token is received on the source chain:

```golang
// UnprefixDenom removes the first portID/channelID pair from a given denomination trace info and returns the
// denomination with the updated trace hash and the new trace info.
// An error is returned if the trace cannot be found on the store from the denom's trace hash.
func (k Keeper) UnprefixDenom(ctx Context, denom string) (denom, trace string, err error) {
  denomSplit := strings.SplitN(denom, "/", 2)

  switch {
    case len(denomSplit) == 0:
      err = Wrap(ErrInvalidDenomForTransfer, denom)
    case denomSplit[0] == denom:
      err = Wrapf(ErrInvalidDenomForTransfer, "denomination should be prefixed with the format 'ibc/{hash(trace + \"/\" + %s)}'", denom)
    case denomSplit[0] != "ibc":
      err = Wrapf(ErrInvalidDenomForTransfer, "denomination %s must start with 'ibc'", denom)
  }

  if err != nil {
    return "", err
  }

  traceHash := tmbytes.HexBytes(denomSplit[1])
  // Get the value from the map trace hash -> trace info prefix
  denomTrace, found = k.GetDenomTrace(ctx, traceHash)
  if !found {
    return "", Wrapf(ErrDenomTraceNotFound, "denom: %s", denom)
  }

  traceSplit := strings.SplitN(denom, "/", 2)
  if len(traceSplit) == 2 {
    // the trace has only one portID/channelID pair
    return denomTrace.BaseDenom
  }

  // remove a single identifiers pair to create the new trace
  denomTrace.Trace = denom.Trace[2:]
  traceHash = denomTrace.Hash()

  // set the value to the lookup table if not stored already
  if !k.HasDenomTrace(ctx, traceHash) {
    k.SetDenomTrace(ctx, traceHash, denomTrace)
  }

  denom = "ibc/"+ traceHash.String()
  return denom
}
```

```golang
// GetTraceFromDenom returns the token source tracing info from the trace hash of the given denomination
func (k Keeper) GetTraceFromDenom(ctx Context, denom string) (string, error) {
  denomSplit := strings.Split(denom, "/")

  if len(denomSplit) == 0 {
    return "", Wrap(ErrInvalidDenomForTransfer, denom)
  } else if denomSplit[0] == denom {
    return "", nil
  }

  traceHash := tmbytes.HexBytes(denomSplit[1])
  // Get the value from the map trace hash -> trace info prefix
  trace, found := k.GetDenomTrace(ctx, traceHash)
  if !found {
    return "", Wrapf(ErrDenomTraceNotFound, "denom: %s", denom)
  }

  return trace
}

```

Additionally, the `SendTransfer`'s `createOutgoingPacket` call and the `OnRecvPacket` need to be
updated to be retreive the trace info (using `GetTraceFromDenom`) prior to checking the correctness
of the prefix.

### Coin Changes

The coin denomination validation will need to be updated to reflect these changes. In particular, the denomination validation
function will now accept slash separators (`"/"`) and will bump the maximum character length to 64.

Additional validation logic, such as verifying the kenght of the hash, the  can be integrated if [custom base denomination validation](https://github.com/cosmos/cosmos-sdk/pull/6755) is integrated into the SDK.

### Positive

- Clearer separation of the source tracing behaviour of the token (transfer prefix) from the original
  `Coin` denomination
- Consistent validation of `Coin` fields (i.e no special characters, fixed max length)
- Cleaner `Coin` and standard denominations for IBC
- No additional fields to SDK `Coin`

### Negative

- Store each set of tracing denomination identifiers on the `ibc-transfer` module store.
- Clients will have to fetch the base denomination everytime they receive a new relayed fungible token over IBC. This can be mitigated using a map for already seen hashes on the client side.

### Neutral

- Slight difference with the ICS20 spec
- Additional validation logic for IBC coins
- Additional genesis fields.
- Slightly increases the gas usage on cross-chain transfers due to access to the store. This should
  be inter-block cached if transfers are frequent

## References

- [ICS 20 - Fungible token transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer)
- [Custom Coin Denomination validation](https://github.com/cosmos/cosmos-sdk/pull/6755)
