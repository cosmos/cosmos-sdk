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

The Packet relay sending works based in 2 cases (per
[specification](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay) and [Colin Axnér](https://github.com/colin-axner)'s description):

1. Sender chain is acting as the source zone. The coins are transferred
to an escrow address (i.e locked) on the sender chain and then transferred
to the receiving chain through IBC TAO logic. It is expected that the
receiving chain will mint vouchers to the receiving address.

2. Sender chain is acting as the sink zone. The coins (vouchers) are burned
on the sender chain and then transferred to the receiving chain though IBC
TAO logic. It is expected that the receiving chain, which had previously
sent the original denomination, will unescrow the fungible token and send
it to the receiving address.

Another way of thinking of source and sink zones is through the token's
timeline. Each send to any chain other than the one it was previously
received from is a movement forwards in the token's timeline. This causes
trace to be added to the token's history and the destination port and
destination channel to be prefixed to the denomination. In these instances
the sender chain is acting as the source zone. When the token is sent back
to the chain it previously received from, the prefix is removed. This is
a backwards movement in the token's timeline and the sender chain
is acting as the sink zone.

### Example

These steps of transfer occur: `A -> B -> C -> A -> C`

1. `A -> B` : sender chain is source zone. Denom upon receiving: `A/denom`
2. `B -> C` : sender chain is source zone. Denom upon receiving: `B/A/denom`
3. `C -> A` : sender chain is source zone. Denom upon receiving: `A/C/B/A/denom`
4. `A -> C` : sender chain is sink zone

The token has a final denomination of `C/B/A/denom`, where `C/B/A` is the trace information.

In this context, when the sender of a cross-chain transfer *is* the source where the tokens
were originated, the protocol prefixes the denomination with the port and channel identifiers in the
following format:

```typescript
prefix + denom = {destPortN}/{destChannelN}/.../{destPort0}/{destChannel0}/denom
```

Example: transferring `100 uatom` from port `HubPort` and channel `HubChannel` on the Hub to
Ethermint's port `EthermintPort` and channel `EthermintChannel` results in `100
EthermintPort/EthermintChannel/uatom`, where `EthermintPort/EthermintChannel/uatom` is the new
denomination on the receiving chain.

In the case those tokens are transferred back to the Hub (i.e the **source** chain), the prefix is
trimmed and the token denomination updated to the original one.

### Problem

The problem of adding additional information to the coin denomination is twofold:

1. The ever increasing length if tokens are transferred to zones other than the source:

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
where only lowercase alphanumeric characters are accepted. While this is desirable for native denominations
to keep a clean UX, it presents a challenge for IBC as ports and channels might be randomly
generated with special and uppercase characters as per the [ICS 024 - Host
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
// DenomTrace contains the base denomination for ICS20 fungible tokens and the source tracing
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
  // NOTE: other cases are checked during msg validation
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

In order to retrieve the trace information from an IBC denomination, a lookup table needs to be
added to the `ibc-transfer` module. These values need to also be persisted between upgrades, meaning
that a new `[]DenomTrace` `GenesisState` field state needs to be added to the module:

```golang
// GetDenom retrieves the full identifiers trace and base denomination from the store.
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
To mitigate this a new `DenomTrace` field needs to be added to the `FungibleTokenPacketData`:

```protobuf
message FungibleTokenPacketData {
  // the token denomination to be transferred
  string denom = 1;
  // the token amount to be transferred
  uint64 amount = 2;
  // coins denomination trace for tracking the source
  DenomTrace denom_trace = 3;
  // the sender address
  string sender = 4;
  // the recipient address on the destination chain
  string receiver = 5;
}
```

The `MsgTransfer` will validate that the `Coin` denomination from the `Token` field contains a valid hash:

```golang
func (msg MsgTransfer) ValidateBasic() error {
  // ...
  if err := msg.Trace.Validate(); err != nil {
    return err
  }
  // Only validate the ibc denomination when trace info is provided
  if msg.Trace.Trace != "" {
    denomTraceHash, err := ValidateIBCDenom(msg.Token.Denom)
    if err != nil {
      return err
    }
    traceHash := msg.Trace.Hash()
    if !bytes.Equal(traceHash.Bytes(), denomTraceHash.Bytes()) {
      return Errorf("token denomination trace hash mismatch, expected %s got %s", traceHash, denomTraceHash)
    }
  } else if msg.Trace.BaseDenom != msg.Token.Denom {
    // otherwise, validate that base denominations are equal
    return Wrapf(
      ErrInvalidDenomForTransfer,
      "token denom must match the trace base denom (%s ≠ %s)",
      msg.Token.Denom, msg.Trace.BaseDenom,
    )
  }
  // ...
}
```

```golang
// ValidateIBCDenom checks that the denomination for an IBC fungible token is valid. It returns the hash of denomination on success.
func ValidateIBCDenom(rawDenom string) (tmbytes.HexBytes, error) {
  denomSplit := strings.SplitN(denom, "/", 2)

  switch {
    case denomSplit[0] == denom:
      err = Wrapf(ErrInvalidDenomForTransfer, "denomination should be prefixed with the format 'ibc/{hash(trace + \"/\" + %s)}'", denom)
    case denomSplit[0] != "ibc":
      err = Wrapf(ErrInvalidDenomForTransfer, "denomination %s must be prefixed with 'ibc'", denom)
    case len(denomSplit) == 2:
      err = tmtypes.ValidateHash([]byte(denomSplit[1]))
    default:
      err = Wrap(ErrInvalidDenomForTransfer, denom)
  }

  if err != nil {
    return nil, err
  }

  return denomSplit[1], nil
}
```

When a fungible token is sent to from a source chain, the trace information needs to be stored with the
new port and channel identifiers:

```golang
// SendTransfer
// ...

// NOTE: SendTransfer simply sends the denomination as it exists on its own
// chain inside the packet data. The receiving chain will perform denom
// prefixing as necessary.

if types.SenderChainIsSource(sourcePort, sourceChannel, token.Denom) {
  // create the escrow address for the tokens
  escrowAddress := types.GetEscrowAddress(sourcePort, sourceChannel)

  // escrow source tokens. It fails if balance insufficient.
  if err := k.bankKeeper.SendCoins(
    ctx, sender, escrowAddress, sdk.NewCoins(token),
  ); err != nil {
    return err
  }
} else {
  // set the value to the lookup table if not stored already
  traceHash := denomTrace.Hash()
  if !k.HasDenomTrace(ctx, traceHash) {
    k.SetDenomTrace(ctx, traceHash, denomTrace)
  }

  // transfer the coins to the module account and burn them
  if err := k.bankKeeper.SendCoinsFromAccountToModule(
    ctx, sender, types.ModuleName, sdk.NewCoins(token),
  ); err != nil {
    return err
  }

  if err := k.bankKeeper.BurnCoins(
    ctx, types.ModuleName, sdk.NewCoins(token),
  ); err != nil {
    // NOTE: should not happen as the module account was
    // retrieved on the step above and it has enough balance
    // to burn.
    return err
  }
}
// ...
```

The denomination trace info also needs to be updated when token is received in both cases:

- Sender is **sink** chain: Store the received denomination, i.e in the [example](#example) above,
  during step 4, when chain `C` receives the `A/C/B/A/denom`. As there the trimmed trace info is
  already known by the chain we don't need to store it (i.e `C/B/A/denom`).
- Sender is **source** chain: Store the received info. For example, during step 1, when chain `B` receives `A/denom`.

```golang
// OnRecvPacket
// ...

// This is the prefix that would have been prefixed to the denomination
// on sender chain IF and only if the token originally came from the
// receiving chain.
//
// NOTE: We use SourcePort and SourceChannel here, because the counterparty
// chain would have prefixed with DestPort and DestChannel when originally
// receiving this coin as seen in the "sender chain is the source" condition.
voucherPrefix := types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())

if types.ReceiverChainIsSource(voucherPrefix, data.Denom) {
  // sender chain is not the source, unescrow tokens

  // set the value to the lookup table if not stored already
  traceHash := denomTrace.Hash()
  if !k.HasDenomTrace(ctx, traceHash) {
    k.SetDenomTrace(ctx, traceHash, denomTrace)
  }

  // remove prefix added by sender chain
  if err := denomTrace.RemovePrefix(); err != nil {
    return err
  }

  // NOTE: since the sender is a sink chain, we already know the unprefixed denomination trace info

  token := sdk.NewCoin(denomTrace.IBCDenom(), sdk.NewIntFromUint64(data.Amount))

  // unescrow tokens
  escrowAddress := types.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
  return k.bankKeeper.SendCoins(ctx, escrowAddress, receiver, sdk.NewCoins(token))
}

// sender chain is the source, mint vouchers

// since SendPacket did not prefix the denomination, we must prefix denomination here
denomTrace.AddPrefix(packet.GetDestPort(), packet.GetDestChannel())

// set the value to the lookup table if not stored already
traceHash := denomTrace.Hash()
if !k.HasDenomTrace(ctx, traceHash) {
  k.SetDenomTrace(ctx, traceHash, denomTrace)
}

voucher := sdk.NewCoin(denomTrace.IBCDenom(), sdk.NewIntFromUint64(data.Amount))

// mint new tokens if the source of the transfer is the same chain
if err := k.bankKeeper.MintCoins(
  ctx, types.ModuleName, sdk.NewCoins(voucher),
); err != nil {
  return err
}

// send to receiver
return k.bankKeeper.SendCoinsFromModuleToAccount(
  ctx, types.ModuleName, receiver, sdk.NewCoins(voucher),
)
```

### Coin Changes

The coin denomination validation will need to be updated to reflect these changes. In particular, the denomination validation
function will now:

- Accept slash separators (`"/"`) and uppercase characters (due to the `HexBytes` format)
- Bump the maximum character length to 64

Additional validation logic, such as verifying the length of the hash, the  may be added to the bank module in the future if the [custom base denomination validation](https://github.com/cosmos/cosmos-sdk/pull/6755) is integrated into the SDK.

### Positive

- Clearer separation of the source tracing behaviour of the token (transfer prefix) from the original
  `Coin` denomination
- Consistent validation of `Coin` fields (i.e no special characters, fixed max length)
- Cleaner `Coin` and standard denominations for IBC
- No additional fields to SDK `Coin`

### Negative

- Store each set of tracing denomination identifiers on the `ibc-transfer` module store
- Clients will have to fetch the base denomination every time they receive a new relayed fungible token over IBC. This can be mitigated using a map/cache for already seen hashes on the client side. Other forms of mitigation, would be opening a websocket connection subscribe to incoming events.

### Neutral

- Slight difference with the ICS20 spec
- Additional validation logic for IBC coins on the `ibc-transfer` module
- Additional genesis fields
- Slightly increases the gas usage on cross-chain transfers due to access to the store. This should
  be inter-block cached if transfers are frequent.

## References

- [ICS 20 - Fungible token transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer)
- [Custom Coin Denomination validation](https://github.com/cosmos/cosmos-sdk/pull/6755)
