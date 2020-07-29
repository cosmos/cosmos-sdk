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
  // chain of port/channel identifiers used for tracing the source of the fungible token
  string trace = 1;
  // base denomination of the relayed fungible token
  string base_denom = 2;
}
```

The `IBCDenom` function constructs the `Coin` denomination used when creating the ICS20 fungible token packet data:

```golang
// Hash returns the hex bytes of the SHA256 hash of the DenomTrace fields.
func (dt DenomTrace) Hash() tmbytes.HexBytes {
  return tmhash.Sum(dt.Trace + "/" + dt.BaseDenom)
}

// IBCDenom a coin denomination for an ICS20 fungible token in the format 'ibc/{hash(trace + baseDenom)}'. If the trace is empty, it will return the base denomination.
func (dt DenomTrace) IBCDenom() string {
  if dt.Trace != "" {
    return fmt.Sprintf("ibc/%s", dt.Hash())
  }
  return dt.BaseDenom
}
```

In order to trim the denomination trace prefix when sending/receiving fungible tokens, the `RemovePrefix` function is provided.

> NOTE: the prefix addition must be done on the client side.

```golang
// RemovePrefix trims the first portID/channelID pair from the trace info. If the trace is already empty it will perform a no-op. If the trace is incorrectly constructed or doesn't have separators it will return an error.
func (dt *DenomTrace) RemovePrefix() error {
  if dt.Trace == "" {
    return nil
  }

  traceSplit := strings.SplitN(dt.Trace, "/", 3)

  var err error
  switch {
  case len(traceSplit) == 0, traceSplit[0] == dt.Trace:
    err = Wrapf(ErrInvalidDenomForTransfer, "trace info %s must contain '/' separators", dt.Trace)
  case len(traceSplit) == 1:
    err = Wrapf(ErrInvalidDenomForTransfer, "trace info %s must come in pairs of '{portID}/channelID}'", dt.Trace)
  case len(traceSplit) == 2:
    dt.Trace = ""
  case len(traceSplit) == 3:
    dt.Trace = traceSplit[2]
  }

  if err != nil {
    return err
  }

  return nil
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
func (k Keeper) SetTrace(ctx Context, denomTrace DenomTrace) {
  store := ctx.KVStore(k.storeKey)
  bz := k.cdc.MustMarshalBinaryBare(&denomTrace)
  store.Set(types.KeyTrace(denomTrace.Hash()), bz)
}
```

The problem with this approach is that when a token is received for the first time, the full trace
info will need to be passed in order to construct the hash and set it to the mapping on the store.
To mitigate this a new `Trace` field needs to be added to the `FungibleTokenPacketData`:

```protobuf
message FungibleTokenPacketData {
  // the tokens to be transferred
  repeated cosmos.Coin amount = 1 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"
  ];
  // coins denomination trace for tracking the source
  DenomTrace denom_trace = 2;
  // the sender address
  string sender = 3;
  // the recipient address on the destination chain
  string receiver = 4;
}
```

The `MsgTransfer` will validate that the Coins from the `Amount` field contain the hash that is valid:

```golang
func (msg MsgTransfer) ValidateBasic() error {
  // ...
  if len(msg.Amount) != len(msg.DenomTraces) {
    //  error
  }
  // tokens length should be 1
  hash, err := GetTraceHashFromDenom(msg.Amount[0].Denom)
  if err != nil {
    return err
  }
  if hash != msg.DenomTrace.Hash() {
    // error
  }
}
```

```golang
func GetTraceHashFromDenom(rawDenom string) (tmbytes.HexBytes, error) {
  denomSplit := strings.SplitN(denom, "/", 3)

  switch {
    case len(denomSplit) == 0:
      err = Wrap(ErrInvalidDenomForTransfer, denom)
    case denomSplit[0] == denom:
      err = Wrapf(ErrInvalidDenomForTransfer, "denomination should be prefixed with the format 'ibc/{hash(trace + \"/\" + %s)}'", denom)
    case denomSplit[0] != "ibc":
      err = Wrapf(ErrInvalidDenomForTransfer, "denomination %s must start with 'ibc'", denom)
    case len(denomSplit) == 2  && len(denomSplit[1]) != 32:
      err = Wrapf(ErrInvalidDenomForTransfer, "invalid SHA256 hash %s", denomSplit[1])
    default:
      err = Wrap(ErrInvalidDenomForTransfer, denom)
  }

  if err != nil {
    return nil, err
  }

  return denomSplit[1], nil
}
```

When a fungible token is sent to a sink chain, the trace information needs to be updated with the
new port and channel identifiers:

The denomination also needs to be updated when token is received on the source chain. This is done during the `SendTransfer`'s `createOutgoingPacket` call and the `OnRecvPacket` call. Below is the updated function that sets the :

```golang
// createOutgoingPacket
// ...
prefix := GetDenomPrefix(destinationPort, destinationChannel)
source := strings.HasPrefix(amount[0].Denom, prefix)

// set the value to the lookup table if not stored already
traceHash := denomTrace.Hash()
if !k.HasDenomTrace(ctx, traceHash) {
  k.SetDenomTrace(ctx, traceHash, denomTrace)
}

if source {
  // clear the denomination from the prefix to send the coins to the escrow account
  coins := make(sdk.Coins, len(amount))
  for i, coin := range amount {
    if strings.HasPrefix(denomTrace.Trace, prefix) {
      if err := denomTrace.RemovePrefix(); err != nil {
        return err
      }

      // set the new value to the lookup table if not stored already
      traceHash := denomTrace.Hash()
      if !k.HasDenomTrace(ctx, traceHash) {
        k.SetDenomTrace(ctx, traceHash, denomTrace)
      }

      coins[i] = sdk.NewCoin(denomTrace.IBCDenom(), coin.Amount)
    } else {
      coins[i] = coin
    }
  }

  // escrow tokens if the destination chain is the same as the sender's
  escrowAddress := GetEscrowAddress(sourcePort, sourceChannel)

  // escrow source tokens. It fails if balance insufficient.
  if err := k.bankKeeper.SendCoins(
    ctx, sender, escrowAddress, coins,
  ); err != nil {
    return err
  }
} else {
  // build the receiving denomination prefix if it's not present
  prefix = GetDenomPrefix(sourcePort, sourceChannel)
  for _, coin := range amount {
    if !strings.HasPrefix(denomTrace.Trace, prefix) {
      return Wrapf(ErrInvalidDenomForTransfer, "invalid token denom trace prefix: %s", denomTrace.IBCDenom())
    }
  }
// ...
}
```

```golang
// OnRecvPacket
// ...
prefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
source := strings.HasPrefix(data.DenomTrace.Trace, prefix)

// ...
// set the value to the lookup table if not stored already
traceHash := denomTrace.Hash()
if !k.HasDenomTrace(ctx, traceHash) {
  k.SetDenomTrace(ctx, traceHash, denomTrace)
}

if source {
  // ... (no changes)
}

// check the denom prefix
prefix = types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
coins := make(sdk.Coins, len(data.Amount))
for i, coin := range data.Amount {
  if !strings.HasPrefix(denomTrace.Trace, prefix) {
    return Wrapf(types.ErrInvalidDenomForTransfer, "invalid token denom trace prefix: %s", denomTrace.IBCDenom(),
    )
  }

  if err := denomTrace.RemovePrefix(); err != nil {
    return err
  }

  // set the new value to the lookup table if not stored already
  traceHash := denomTrace.Hash()
  if !k.HasDenomTrace(ctx, traceHash) {
    k.SetDenomTrace(ctx, traceHash, denomTrace)
  }

  coins[i] = sdk.NewCoin(denomTrace.IBCDenom(), coin.Amount)
}
// ...
```

### Coin Changes

The coin denomination validation will need to be updated to reflect these changes. In particular, the denomination validation
function will now accept slash separators (`"/"`) and will bump the maximum character length to 64.

Additional validation logic, such as verifying the kenght of the hash, the  may be added to the bank module in the future if the [custom base denomination validation](https://github.com/cosmos/cosmos-sdk/pull/6755) is integrated into the SDK.

### Positive

- Clearer separation of the source tracing behaviour of the token (transfer prefix) from the original
  `Coin` denomination
- Consistent validation of `Coin` fields (i.e no special characters, fixed max length)
- Cleaner `Coin` and standard denominations for IBC
- No additional fields to SDK `Coin`

### Negative

- Store each set of tracing denomination identifiers on the `ibc-transfer` module store
- Clients will have to fetch the base denomination everytime they receive a new relayed fungible token over IBC. This can be mitigated using a map/cache for already seen hashes on the client side.

### Neutral

- Slight difference with the ICS20 spec
- Additional validation logic for IBC coins on the `ibc-transfer` module
- Additional genesis fields
- Slightly increases the gas usage on cross-chain transfers due to access to the store. This should
  be inter-block cached if transfers are frequent.

## References

- [ICS 20 - Fungible token transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer)
- [Custom Coin Denomination validation](https://github.com/cosmos/cosmos-sdk/pull/6755)
