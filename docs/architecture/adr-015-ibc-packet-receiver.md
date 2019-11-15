# ADR 015: IBC Packet Receiver

## Changelog

- 2019 Oct 22: Initial Draft

## Context

[ICS 26 - Routing Module](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module) defines function [`handlePacketRecv`](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module#packet-relay).

`handlePacketRecv` executes per-module `onRecvPacket` callbacks, verifies the
packet merkle proof, and pushes the acknowledgement bytes, if present, to the IBC
channel `Keeper` state (ICS04).

`handlePacketAcknowledgement` executes per-module `onAcknowledgementPacket`
callbacks, and verifies the acknowledgement commitment proof.

`handlePacketTimeout` and `handlePacketTimeoutOnClose` executes per-module
`onTimeoutPacket` callbacks, and verifies the timeout proof.

The mechanism is similar to the transaction handling logic in `baseapp`. After
authentication, the handler is executed, and the authentication state change
must be committed regardless of the result of the handler execution.

`handlePacketRecv` also requires acknowledgement writing which has to be done
after the handler execution and also must be commited regardless of the result of
the handler execution.

## Decision

The Cosmos SDK will define an `AnteDecorator` for IBC packet receiving. The
`AnteDecorator` will iterate over the messages included in the transaction, type
`switch` to check whether the message contains an incoming IBC packet, and if so
verify the Merkle proof.

```go
// Pseudocode
type ProofVerificationDecorator struct {
  clientKeeper ClientKeeper
  channelKeeper ChannelKeeper
}

func (pvr ProofVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
  for _, msg := range tx.GetMsgs() {
    var err error
    switch msg := msg.(type) {
    case client.MsgUpdateClient:
      err = pvr.clientKeeper.UpdateClient(msg.ClientID, msg.Header)
    case channel.MsgPacket:
      err = pvr.channelKeeper.VerifyPacket(msg.Packet, msg.Proofs, msg.ProofHeight)
      // Store the empty acknowledgement for convinience
      pvr.channelKeeper.SetPacketAcknowledgement(ctx, msg.PortID, msg.ChannelID, msg.Sequence, []byte{})
    case chanel.MsgAcknowledgement:
      err = pvr.channelKeeper.VerifyAcknowledgement(msg.Acknowledgement, msg.Proof, msg.ProofHeight)
    case channel.MsgTimeoutPacket:
      err = pvr.channelKeeper.VerifyTimeout(msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv)
    default:
      continue
    }

    if err != nil {
      return ctx, err
    }
  }
  return next(ctx, tx, simulate)
}
```

Where `MsgUpdateClient`, `MsgPacket`, `MsgAcknowledgement`, `MsgTimeoutPacket`
are `sdk.Msg` types correspond to `handleUpdateClient`, `handleRecvPacket`,
`handleAcknowledgementPacket`, `handleTimeoutPacket` of the routing module,
respectively.

The `ProofVerificationDecorator` will be inserted to the top level application.
It should come right after the default sybil attack resistent layer from the
current `auth.NewAnteHandler`:

```go
// add IBC ProofVerificationDecorator to the Chain of
func NewAnteHandler(
  ak keeper.AccountKeeper, supplyKeeper types.SupplyKeeper, ibcKeeper ibc.Keeper,
  sigGasConsumer SignatureVerificationGasConsumer) sdk.AnteHandler {
  return sdk.ChainAnteDecorators(
    NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
    ...
    NewIncrementSequenceDecorator(ak),
    ibcante.ProofVerificationDecorator(ibcKeeper.ClientKeeper, ibcKeeper.ChannelKeeper), // innermost AnteDecorator
  )
}
```

The Cosmos SDK will define the wrapper function `WriteAcknowledgement` under the
ICS05 port keeper. The function will wrap packet handlers to automatically handle
the acknowledgments.

```go
type PacketHandler func(sdk.Context, Packet) sdk.Result

func (k PortKeeper) WriteAcknowledgement(ctx sdk.Context, msg MsgPacket, h PacketHandler) sdk.Result {
  // Cache context
  cacheCtx, write := ctx.CacheContext()

  // verification already done inside the antehandler
  res := h(cacheCtx, msg.Packet)
  
  // write the cache only if succedded
  if res.IsOK() {
    write()
  }
  
  // set the result to OK to persist the state change
  res.Code = sdk.CodeOK
  
  // res.Data will be stored as acknowledgement; noop if not exists(empty bytes already stored)
  if res.Data != nil {
    k.SetPacketAcknowledgement(ctx, msg.PortID, msg.ChannelID, msg.Sequence, res.Data)
  }

  return res
}
```

Example application-side usage:

```go
func NewHandler(k Keeper) sdk.Handler {
  return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
    switch msg := msg.(type) {
    case ibc.MsgPacket:
      return k.port.WriteAcknowledgement(ctx, msg, func(ctx sdk.Context, p ibc.Packet) sdk.Result {
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

func handleCustomPacket(ctx sdk.Context, k Keeper, packet MyPacket) sdk.Result {
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

### Negative

- Cannot support dynamic ports, routing is tied to the baseapp router
  Dynamic ports can be supported using hierarchical port identifier, see #5290 for detail

### Neutral

- Introduces new `AnteHandler` decorator.

## References

- Relevant comment: [cosmos/ics#289](https://github.com/cosmos/ics/issues/289#issuecomment-544533583)
- [ICS26 - Routing Module](https://github.com/cosmos/ics/blob/master/spec/ics-026-routing-module)
