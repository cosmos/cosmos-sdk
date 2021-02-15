<!--
order: 4
-->

# Messages

## MsgTransfer

A fungible token cross chain transfer is achieved by using the `MsgTransfer`:

```go
type MsgTransfer struct {
  SourcePort        string
  SourceChannel     string
  Token             sdk.Coin
  Sender            string
  Receiver          string
  TimeoutHeight     ibcexported.Height
  TimeoutTimestamp  uint64
}
```

This message is expected to fail if:

- `SourcePort` is invalid (see 24-host naming requirements)
- `SourceChannel` is invalid (see 24-host naming requirements)
- `Token` is invalid (denom is invalid or amount is negative)
- `Token.Amount` is not positive
- `Sender` is empty
- `Receiver` is empty
- `TimeoutHeight` and `TimeoutTimestamp` are both zero
- `Token.Denom` is not a valid IBC denomination as per [ADR 001 - Coin Source Tracing](./../../../../../docs/architecture/adr-001-coin-source-tracing.md).

This message will send a fungible token to the counterparty chain represented
by the counterparty Channel End connected to the Channel End with the identifiers
`SourcePort` and `SourceChannel`.

The denomination provided for transfer should correspond to the same denomination
represented on this chain. The prefixes will be added as necessary upon by the
receiving chain.
