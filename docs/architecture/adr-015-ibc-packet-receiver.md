# ADR 015: IBC Packet Receiver

## Changelog

- 2019 Oct 22: Initial Draft

## Context

[ICS 26](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module) defines function [`handlePacketRecv`](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module#packet-relay). 
`handlePacketRecv` executes per-module `onRecvPacket` callbacks, verifies the packet merkle proof, and pushes the acknowledgement bytes, if present,
to the state. 

The mechanism is similar to the transaction handling logic in `baseapp`. After authentication, the handler is executed, and 
the authentication state change must be committed regardless of the result of the handler execution. 

ICS 26 also requires acknowledgement writing which has to be done after the handler execution and also must be commited 
regardless of the result of the handler execution.

## Decision

We will define an `AnteHandler` for IBC packet receiving. The `AnteHandler` will iterate over the messages included in the 
transaction, type switch to check whether the message contains an incoming IBC packet, and if so verify the Merkle proof.

```go
// Pseudocode
func IBCAnteHandler(ctx sdk.Context, tx sdk.Tx, bool) (sdk.Context, sdk.Result, bool) {
  for _, msg := range tx.GetMsgs() {
    if msg, ok := msg.(MsgPacket); ok {
      if err := VerifyPacket(msg.Packet, msg.Proofs, msg.ProofHeight); err != nil {
         return ctx, err.Result(), true
      }
    }
  }
  return ctx, sdk.Result{}, false
}
```

where `MsgPacket` is the `sdk.Msg` type including any IBC packet inside and embedding `Packet.Route()` method.

The `AnteHandler` will be inserted to the top level application, after the signature authentication logic provided by `auth.NewAnteHandler`, utilizing `AnteDecorator` pattern.

This proposal does not cover automatic acknowledgement processing. Applications who want to return acknowledgement should manually handle multistore cacheing.

## Status

Proposed

## Consequences

### Positive

- Intuitive interface for developers - IBC handlers do not need to care about IBC authentication
- State change commitment logic is embedded into `baseapp.runTx` logic
- IBC applications do not need to define message type for receiving packet

### Negative

- AnteHandler processes over the whole transaction before the each handler, `UpdateClient` and `RecvPacket` cannot be processed in the same transaction
- Cannot support dynamic ports, routing is tied to the baseapp router

### Neutral

- Introduces new AnteHandler

## References

- Relevant comment: https://github.com/cosmos/ics/issues/289#issuecomment-544533583
