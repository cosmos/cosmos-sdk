# ADR 015: IBC Packet Receiver

## Changelog

- 2019 Oct 22: Initial Draft

## Context

[ICS 26](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module) defines function [`handlePacketRecv`](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module#packet-relay). 
`handlePacketRecv` executes per-module `onRecvPacket` callbacks, verifies the packet merkle proof, and pushes the acknowledgement bytes, if present,
to the IBC channel `Keeper` state (ICS04). 
`handlePacketAcknowledgement` executes per-module `onAcknowledgementPacket` callbacks, and verifies the acknowledgement merkle proof.
`handlePacketTimeout` and `handlePacketTimeoutOnClose` executes per-module `onTimeoutPacket` callbacks, and verifies the timeout proof.

The mechanism is similar to the transaction handling logic in `baseapp`. After authentication, the handler is executed, and 
the authentication state change must be committed regardless of the result of the handler execution. 

`handlePacketRecv` also requires acknowledgement writing which has to be done after the handler execution and also must be commited 
regardless of the result of the handler execution.

## Decision

The Cosmos SDK will define an `AnteHandler` for IBC packet receiving. The `AnteHandler` will iterate over the messages included in the 
transaction, type switch to check whether the message contains an incoming IBC packet, and if so verify the Merkle proof.

```go
// Pseudocode
func IBCAnteHandler(ctx sdk.Context, tx sdk.Tx, bool) (sdk.Context, sdk.Result, bool) {
  for _, msg := range tx.GetMsgs() {
    switch msg := msg.(type) {
    case ibc.MsgUpdateClient:
      if err := UpdateClient(msg.ClientID, msg.Header); err != nil {
        return ctx, err.Result(), true
      }
    case ibc.MsgPacket:
      if err := VerifyPacket(msg.Packet, msg.Proofs, msg.ProofHeight); err != nil {
        return ctx, err.Result(), true
      }
    case ibc.MsgAcknowledgement:
      if err := VerifyAcknowledgement(msg.Acknowledgement, msg.Proof, msg.ProofHeight); err != nil {
        return ctx, err.Result(), true
      }
    case ibc.MsgTimeoutPacket:
      if err := VerifyTimeout(msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv); err != nil {
        return ctx, err.Result(), true
      }
    }
  }
  return ctx, sdk.Result{}, false
}
```

where `MsgUpdateClient`, `MsgPacket`, `MsgAcknowledgement`, `MsgTimeoutPacket` are `sdk.Msg` types invoking `handleUpdateClient`, `handleRecvPacket`, `handleAcknowledgementPacket`, `handleTimeoutPacket` of the routing module, respectively.

The `AnteHandler` will be inserted to the top level application, after the signature authentication logic provided by `auth.NewAnteHandler`, utilizing `AnteDecorator` pattern.

The Cosmos SDK will define the wrapper function `ReceivePacket` under the ICS05 port keeper. The function will wrap packet handlers to automatically handle the acknowledgments.

```go
// Pseudocode
type Ack struct {
  Data []byte
  Err sdk.Error
}

func (ack Ack) IsOK() bool {
  return ack.Err.IsOK()
}

func (k PortKeeper) ReceivePacket(ctx sdk.Context, msg MsgPacket, h func(sdk.Context, Packet) ibc.Ack) sdk.Result {
  // Cache context
  cacheCtx, write := ctx.CacheContext()

  // verification already done inside the antehandler
  ack := h(cacheCtx, msg.Packet)
  
  // write the cache only if succedded
  if ack.IsOK() {
    write()
  }
  
  res := ack.Err.Result()
  // set the result to OK to persist the state change
  res.Code = sdk.CodeOK
  
  // ackData will be stored as acknowledgement; []byte{} will be stored if not exists
  if ack.Data == nil {
    ack.Data = []byte{}
  }
  k.SetPacketAcknowledgement(ctx, msg.PortID, msg.ChannelID, msg.Sequence, ack.Data)

  return res
}
```

Example application-side usage:

```go
func NewHandler(k Keeper) sdk.Handler {
  return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
    switch msg := msg.(type) {
    case ibc.MsgPacket:
      return k.port.ReceivePacket(ctx, msg, func(ctx sdk.Context, p Packet) ibc.Ack {
        switch packet := packet.(type) {
        case CustomPacket: // i.e fulfills the Packet interface
          return handleCustomPacket(ctx, k, packet)
        }
      })
    case ibc.MsgAcknowledgement:
      switch ack := msg.Acknowledgement.(type) {
      case CustomAcknowledgement:
        return handleCustomAcknowledgement(ctx, k, msg.Acknowledgement)
      }
    case ibc.MsgTimeoutPacket:
      switch packet := msg.Packet.(type) {
      case CustomPacket:
        return handleCustomTimeoutPacket(ctx, k, msg.Packet)
      }
    }
  }
}

func handleCustomPacket(ctx sdk.Context, k Keeper, packet MyPacket) ibc.Ack {
  if failureCondition {
    return AckInvalidPacketContent(k.codespace, []byte{packet.Data})
  }
  // Handler logic
  return sdk.Result{}
}

func handleCustomAcknowledgement(ctx sdk.Context, k Keeper, ack MyAcknowledgement) (res sdk.Result) {
  // Handler logic
  return
}

func handleCustomTimeoutPacket(ctx sdk.Context, k Keeper, packet MyPacket) (res sdk.Result) {
  // Handler logic
  return
}
```

## Status

Proposed

## Consequences

### Positive

- Intuitive interface for developers - IBC handlers do not need to care about IBC authentication
- State change commitment logic is embedded into `baseapp.runTx` logic
- IBC applications do not need to define message type for receiving packet

### Negative

- Cannot support dynamic ports, routing is tied to the baseapp router

### Neutral

- Introduces new AnteHandler

## References

- Relevant comment: https://github.com/cosmos/ics/issues/289#issuecomment-544533583
