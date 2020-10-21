# ADR 001: Coin Source Tracing

## Changelog

- 2020-07-09: Initial Draft
- 2020-08-11: Implementation changes

## Status

Accepted, Implemented

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

Assume the following channel connections exist and that all channels use the port ID `transfer`:

- chain `A` has channels with chain `B` and chain `C` with the IDs `channelToB` and `channelToC`, respectively
- chain `B` has channels with chain `A` and chain `C` with the IDs `channelToA` and `channelToC`, respectively
- chain `C` has channels with chain `A` and chain `B` with the IDs `channelToA` and `channelToB`, respectively

These steps of transfer between chains occur in the following order: `A -> B -> C -> A -> C`. In particular:

1. `A -> B`: sender chain is source zone. `A` sends packet with `denom` (escrowed on `A`), `B` receives `denom` and mints and sends voucher `transfer/channelToA/denom` to recipient.
2. `B -> C`: sender chain is source zone. `B` sends packet with `transfer/channelToA/denom` (escrowed on `B`), `C` receives `transfer/channelToA/denom` and mints and sends voucher `transfer/channelToB/transfer/channelToA/denom` to recipient.
3. `C -> A`: sender chain is source zone. `C` sends packet with `transfer/channelToB/transfer/channelToA/denom` (escrowed on `C`), `A` receives `transfer/channelToB/transfer/channelToA/denom` and mints and sends voucher `transfer/channelToC/transfer/channelToB/transfer/channelToA/denom` to recipient.
4. `A -> C`: sender chain is sink zone. `A` sends packet with `transfer/channelToC/transfer/channelToB/transfer/channelToA/denom` (burned on `A`), `C` receives `transfer/channelToC/transfer/channelToB/transfer/channelToA/denom`, and unescrows and sends `transfer/channelToB/transfer/channelToA/denom` to recipient.

The token has a final denomination on chain `C` of `transfer/channelToB/transfer/channelToA/denom`, where `transfer/channelToB/transfer/channelToA` is the trace information.

In this context, upon a receive of a cross-chain fungible token transfer, if the sender chain is the source of the token, the protocol prefixes the denomination with the port and channel identifiers in the following format:

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

The issues outlined above, are applicable only to SDK-based chains, and thus the proposed solution
are do not require specification changes that would result in modification to other implementations
of the ICS20 spec.

Instead of adding the identifiers on the coin denomination directly, the proposed solution hashes
the denomination prefix in order to get a consistent length for all the cross-chain fungible tokens.

This will be used for internal storage only, and when transferred via IBC to a different chain, the
denomination specified on the packed data will be the full prefix path of the identifiers needed to
trace the token back to the originating chain, as specified on ICS20.

The new proposed format will be the following:

```golang
ibcDenom = "ibc/" + hash(trace path + "/" + base denom)
```

The hash function will be a SHA256 hash of the fields of the `DenomTrace`:

```protobuf
// DenomTrace contains the base denomination for ICS20 fungible tokens and the source tracing
// information
message DenomTrace {
  // chain of port/channel identifiers used for tracing the source of the fungible token
  string path = 1;
  // base denomination of the relayed fungible token
  string base_denom = 2;
}
```

The `IBCDenom` function constructs the `Coin` denomination used when creating the ICS20 fungible token packet data:

```golang
// Hash returns the hex bytes of the SHA256 hash of the DenomTrace fields using the following formula:
//
// hash = sha256(tracePath + "/" + baseDenom)
func (dt DenomTrace) Hash() tmbytes.HexBytes {
  return tmhash.Sum(dt.Path + "/" + dt.BaseDenom)
}

// IBCDenom a coin denomination for an ICS20 fungible token in the format 'ibc/{hash(tracePath + baseDenom)}'. 
// If the trace is empty, it will return the base denomination.
func (dt DenomTrace) IBCDenom() string {
  if dt.Path != "" {
    return fmt.Sprintf("ibc/%s", dt.Hash())
  }
  return dt.BaseDenom
}
```

### `x/ibc-transfer` Changes

In order to retrieve the trace information from an IBC denomination, a lookup table needs to be
added to the `ibc-transfer` module. These values need to also be persisted between upgrades, meaning
that a new `[]DenomTrace` `GenesisState` field state needs to be added to the module:

```golang
// GetDenomTrace retrieves the full identifiers trace and base denomination from the store.
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

// HasDenomTrace checks if a the key with the given trace hash exists on the store.
func (k Keeper) HasDenomTrace(ctx Context, denomTraceHash []byte)  bool {
  store := ctx.KVStore(k.storeKey)
  return store.Has(types.KeyTrace(denomTraceHash))
}

// SetDenomTrace sets a new {trace hash -> trace} pair to the store.
func (k Keeper) SetDenomTrace(ctx Context, denomTrace DenomTrace) {
  store := ctx.KVStore(k.storeKey)
  bz := k.cdc.MustMarshalBinaryBare(&denomTrace)
  store.Set(types.KeyTrace(denomTrace.Hash()), bz)
}
```

The `MsgTransfer` will validate that the `Coin` denomination from the `Token` field contains a valid
hash, if the trace info is provided, or that the base denominations matches:

```golang
func (msg MsgTransfer) ValidateBasic() error {
  // ...
  return ValidateIBCDenom(msg.Token.Denom)
}
```

```golang
// ValidateIBCDenom validates that the given denomination is either:
//
//  - A valid base denomination (eg: 'uatom')
//  - A valid fungible token representation (i.e 'ibc/{hash}') per ADR 001 https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-001-coin-source-tracing.md
func ValidateIBCDenom(denom string) error {
  denomSplit := strings.SplitN(denom, "/", 2)

  switch {
  case strings.TrimSpace(denom) == "",
    len(denomSplit) == 1 && denomSplit[0] == "ibc",
    len(denomSplit) == 2 && (denomSplit[0] != "ibc" || strings.TrimSpace(denomSplit[1]) == ""):
    return sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "denomination should be prefixed with the format 'ibc/{hash(trace + \"/\" + %s)}'", denom)

  case denomSplit[0] == denom && strings.TrimSpace(denom) != "":
    return sdk.ValidateDenom(denom)
  }

  if _, err := ParseHexHash(denomSplit[1]); err != nil {
    return Wrapf(err, "invalid denom trace hash %s", denomSplit[1])
  }

  return nil
}
```

The denomination trace info only needs to be updated when token is received:

- Receiver is **source** chain: The receiver created the token and must have the trace lookup already stored (if necessary _ie_ native token case wouldn't need a lookup).
- Receiver is **not source** chain: Store the received info. For example, during step 1, when chain `B` receives `transfer/channelToA/denom`.

```golang
// SendTransfer
// ...

  fullDenomPath := token.Denom

// deconstruct the token denomination into the denomination trace info
// to determine if the sender is the source chain
if strings.HasPrefix(token.Denom, "ibc/") {
  fullDenomPath, err = k.DenomPathFromHash(ctx, token.Denom)
  if err != nil {
    return err
  }
}

if types.SenderChainIsSource(sourcePort, sourceChannel, fullDenomPath) {
//...
```

```golang
// DenomPathFromHash returns the full denomination path prefix from an ibc denom with a hash
// component.
func (k Keeper) DenomPathFromHash(ctx sdk.Context, denom string) (string, error) {
  hexHash := denom[4:]
  hash, err := ParseHexHash(hexHash)
  if err != nil {
    return "", Wrap(ErrInvalidDenomForTransfer, err.Error())
  }

  denomTrace, found := k.GetDenomTrace(ctx, hash)
  if !found {
    return "", Wrap(ErrTraceNotFound, hexHash)
  }

  fullDenomPath := denomTrace.GetFullDenomPath()
  return fullDenomPath, nil
}
```


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
if ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
  // sender chain is not the source, unescrow tokens

  // remove prefix added by sender chain
  voucherPrefix := types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
  unprefixedDenom := data.Denom[len(voucherPrefix):]
  token := sdk.NewCoin(unprefixedDenom, sdk.NewIntFromUint64(data.Amount))

  // unescrow tokens
  escrowAddress := types.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
  return k.bankKeeper.SendCoins(ctx, escrowAddress, receiver, sdk.NewCoins(token))
}

// sender chain is the source, mint vouchers

// since SendPacket did not prefix the denomination, we must prefix denomination here
sourcePrefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
// NOTE: sourcePrefix contains the trailing "/"
prefixedDenom := sourcePrefix + data.Denom

// construct the denomination trace from the full raw denomination
denomTrace := types.ParseDenomTrace(prefixedDenom)

// set the value to the lookup table if not stored already
traceHash := denomTrace.Hash()
if !k.HasDenomTrace(ctx, traceHash) {
  k.SetDenomTrace(ctx, traceHash, denomTrace)
}

voucherDenom := denomTrace.IBCDenom()
voucher := sdk.NewCoin(voucherDenom, sdk.NewIntFromUint64(data.Amount))

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

```golang
func NewDenomTraceFromRawDenom(denom string) DenomTrace{
  denomSplit := strings.Split(denom, "/")
  trace := ""
  if len(denomSplit) > 1 {
    trace = strings.Join(denomSplit[:len(denomSplit)-1], "/")
  }
  return DenomTrace{
    BaseDenom: denomSplit[len(denomSplit)-1],
    Trace:     trace,
  }
}
```

One final remark is that the `FungibleTokenPacketData` will remain the same, i.e with the prefixed full denomination, since the receiving chain may not be an SDK-based chain.

### Coin Changes

The coin denomination validation will need to be updated to reflect these changes. In particular, the denomination validation
function will now:

- Accept slash separators (`"/"`) and uppercase characters (due to the `HexBytes` format)
- Bump the maximum character length to 128, as the hex representation used by Tendermint's
  `HexBytes` type contains 64 characters.

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
