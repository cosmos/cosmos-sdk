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
Ethermint's port `EthermintPort` and channel `EthermintChannel` results in `10
EthermintPort/EthermintChannel/uatom`, where `EthermintPort/EthermintChannel/uatom` is the new
denomination on the receiving chain.

In the case those tokens are transfered back to the Hub (i.e the **source** chain), the prefix is
trimmed and the token denomination updated to the original one.

### Problem

The problem of adding additional information to the coin denomination is twofold:

1. The ever increasing length if tokens are transfered to zones other than the source:

If a token is transfered `n` times via IBC to a sink chain, the token denom will contain `n` pairs
of prefixes, as shown on the format example above. This is supposes a problem because, while port
and channel identifiers don't have a cap on the maximum lenght, the SDK `Coin` type only accepts
denoms up to 64 characters.

This can result undesired behaviours such as tokens not being abled to be transfered to multiple
sink chains if the denomination exceeds the lenght or unexpected `panics` due to denomination
validation on the receiving chain.

2. The existence of special characters and upercase letters on the denomination:

In the SDK everytime a `Coin` is initialized trhough the constructor function `NewCoin`, a validation
of a coin's denom is performed according to a
[Regex](https://github.com/cosmos/cosmos-sdk/blob/a940214a4923a3bf9a9161cd14bd3072299cd0c9/types/coin.go#L583),
where only lowercase alphanumeric characters are accepted. While this is desirable for native denoms
to keep a clean UX, it supposes a challenge for IBC as ports and channels might be randomly
generated with special carracters and uppercases as per the [ICS 024 - Host
Requirements](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements#paths-identifiers-separators)
specification.

## Decision

Introduce a new `Trace` field to the SDK's `Coin` type so that the two problems are mitigated.

<!-- TODO: change field to metadata -->

```protobuf
// Coin defines a token with a denomination and an amount.
//
// NOTE: The amount field is an Int which implements the custom method
// signatures required by gogoproto.
message Coin {
  option (gogoproto.equal) = true;

  string denom  = 1;
  string amount = 2 [(gogoproto.customtype) = "Int", (gogoproto.nullable) = false];
  // trace the origin of the token. Every time a Coin is transferred to a chain that's not the souce
  // of the token, a new item is inserted to the first position.
  repeated Trace trace = 3;
}

// Trace defines a origin tracing logic for fungible token cross-chain token transfers through
// IBC (as specified per ICS20).
message Trace {
  // destination chain port identifier
  string port_id = 1 [(gogoproto.moretags) = "yaml:\"port_id\""];
  // destination chain channel identifier
  string channel_id = 2 [(gogoproto.moretags) = "yaml:\"channel_id\""];
}
```

To prevent breaking the `NewCoin` constructor, a separate `NewCoinWithTrace` function will be
created.

```go

// NewCoinWithTrace creates a new coin with .
func NewCoinWithTrace(denom string, amount Int, traces ...Trace) Coin {
  coin := NewCoin(denom, amount)

  for _, trace := range traces {
    if err := validateTrace(trace); err != nil {
      panic(err)
    }
  }

  coin.Trace = traces
  return coin
  }
```

To transfer tokens to a sink chain via IBC, `InsertTrace` can be used:

```go
// InsertTrace validates the destination port and channel identifiers and inserts a Trace
// insance to the first position of the list.
func (coin *Coin) InsertTrace(portID, channelID string) error {
  if err := validatePortID(portID); err != nil {
    return err
  }

  if err := validateChannelID(channelID); err != nil {
    return err
  }

  coin.Trace = append([]Trace{NewTrace(portID, channelID)}, coin.Trace...)
  return nil
  }
```

To delete a trace instance, the `PopTrace` utity function can be used:

```go
// PopTrace removes and returns the first trace item from the list. If the list is empty
// an error is returned instead.
func (coin *Coin) PopTrace() (Trace, error) {
  if len(coin.Trace) == 0 {
    return Trace{}, errors.New("trace list is empty")
  }

  trace := coin.Trace[0]
  coin.Trace = coin.Trace[1:]
  return trace, nil
  }
```

<!-- TODO: updates to ICS20 -->

### Positive

- Clearer separation of the origin tracing behaviour of the token (transfer prefix) from the original
  `Coin` denomination
- Consistent validation of `Coin` fields
- Prevents clients from having to strip the denomination
- Cleaner `Coin` denominations for IBC instead of prefixed ones

### Negative

- `Coin` metadata size is not bounded, which might result in a significant increase of the size per
  `Coin` compared with the current implementation.

### Neutral

- Additional field to the `Coin` type
- Slight difference with the ICS20 spec

## References

- [ICS 20 - Fungible token
  transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer)