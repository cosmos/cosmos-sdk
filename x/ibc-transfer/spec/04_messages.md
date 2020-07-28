<!--
order: 4
-->

# Messages

### MsgTransfer

A fungible token cross chain transfer is achieved by using the `MsgTransfer`:

```go
type MsgTransfer struct {
  SourcePort        string
  SourceChannel     string
  Amount            sdk.Coin
  Sender            sdk.AccAddress
  Receiver          string
  Source            bool
  TimeoutHeight     uint64
  TimeoutTimestamp  uint64
}
```

This message is expected to fail if:

- `SourcePort` is invalid (see 24-host naming requirements)
- `SourceChannel` is invalid (see 24-host naming requirements)
- `Amount` is invalid (denom is invalid or amount is negative)
- `Amount` is not positive
` `Sender` is empty
- `Receiver` is empty
- `TimeoutHeight` and `TimeoutTimestamp` are both zero

This message will send a fungible token to the counterparty chain represented
by the counterparty Channel End connected to the Channel End with the identifiers
`SourcePort` and `SourceChannel`.

The `Source` field indicates whether the token will be escrowed or burnt
on the sending chain. The `Source` field should be false **only if** the 
ics20 token is being sent back to the chain that caused it to be minted. 

The denomination provided for transfer should correspond to the same denomination
represented on this chain. The prefixes needed to send to the counterparty 
chain will be added in protocol. The fungible token packet created in protocol
will contain this prefixed denomination **only if** source is set to true. The
token is adding to the trace (ie being sent to a different chain then it came 
from).


