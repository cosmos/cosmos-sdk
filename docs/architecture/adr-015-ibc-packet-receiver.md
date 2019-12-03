# ADR 015: IBC Packet Receiver

## Changelog

- 2019 Oct 22: Initial Draft

## Context
 
[ICS 26 - Routing Module](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module) defines a function [`handlePacketRecv`](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module#packet-relay).

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

The Cosmos SDK will define `FoldHandler` for post-execution cleanup logic. 

```go
type FoldHandler func(sdk.Context, sdk.Tx, sdk.Result) sdk.Result
```

`FoldHandler`s will be provided by the top level application and interted into
the `baseapp`.

`baseapp.runTx` will execute `FoldHandler` after the main application handler
execution. The logic is equal to that of `AnteHandler`.

```go
// Pseudocode
func (app *BaseApp) runTx(tx sdk.Tx) (result sdk.Result) {
  msgs := tx.GetMsgs()
  
  // AnteHandler
  if app.anteHandler != nil {
    anteCtx, msCache := app.cacheTxContext(ctx)
    newCtx, err := app.anteHandler(anteCtx, tx)
    if !newCtx.IsZero() {
      ctx = newCtx.WithMultiStore(ms)
    }
    
    if err != nil {
      // error handling logic
      return res
    }
    
    msCache.Write()
  }
  
  // Main Handler
  runMsgCtx, msCache := app.cacheTxContext(ctx)
  result = app.runMsgs(runMsgCtx, msgs)
  if !result.IsOK() {
    msCache.Write()
  }
  
  // BEGIN modification made in this ADR
  if app.foldHandler != nil {
    result = app.foldHandler(ctx, tx, result)
  }
  // END
  
  return result
}
```

The Cosmos SDK will define an `AnteDecorator` for IBC packet receiving. The
`AnteDecorator` will iterate over the messages included in the transaction, type
`switch` to check whether the message contains an incoming IBC packet, and if so
verify the Merkle proof.

```go
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
      err = pvr.channelKeeper.RecvPacket(msg.Packet, msg.Proofs, msg.ProofHeight)
    case chanel.MsgAcknowledgement:
      err = pvr.channelKeeper.AcknowledgementPacket(msg.Acknowledgement, msg.Proof, msg.ProofHeight)
    case channel.MsgTimeoutPacket:
      err = pvr.channelKeeper.TimeoutPacket(msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv)
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

The side effects of `RecvPacket`, `VerifyAcknowledgement`, 
`VerifyTimeout` will be extracted out into separated functions. 

```go
// Pseudocode
func (keeper ChannelKeeper) RecvFinalize(packet Packet, proofs []commitment.ProofI, height uint64) {
  keeper.SetNextSequenceRecv(ctx, packet.GetDestPort(), packet.GetDestChannel(), nextSequenceRecv)
}

// Pseudocode
func (keeper ChannelKeeper) AcknowledgementFinalize(packet Packet, acknowledgement PacketDataI, proofs []commitment.ProofI, height uint64) {
  keeper.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
}

// Pseudocode
func (keeper ChannelKeeper) TimeoutFinalize(packet Packet, proofs []commitment.ProofI, height uint64, nextSequenceRecv uint64) {
  k.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
  
  if channel.Ordering == types.ORDERED [
    channel.State = types.CLOSED
    k.SetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), channel)
  }
}
```

The Cosmos SDK will define a `FoldHandler` for remaining state mutation. The
`FoldHandler` will execute the side effect of the verification, including 
sequence increase and commitment deletion.

```go
func NewFoldHandler(k ChannelKeeper) sdk.FoldHandler {
  return func(ctx sdk.Context, tx sdk.Tx, result sdk.Result) sdk.Result {
    if !result.IsOK() {
      // Transaction aborted, no side effect need to be committed
      return result
    }
    
    for _, msg := range tx.GetMsgs() {
      switch msg := msg.(type) {
      case channel.MsgPacket:
        pvr.channelKeeper.FinalizeRecvPacket(msg.Packet, msg.Proofs, msg.ProofHeight)
      case chanel.MsgAcknowledgement:
        pvr.channelKeeper.FinalizeAcknowledgementPacket(msg.Acknowledgement, msg.Proof, msg.ProofHeight)
      case channel.MsgTimeoutPacket:
        pvr.channelKeeper.FinalizeTimeoutPacket(msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv)
      default:
        continue
      }
    }
    
    return result
  }
}
```

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

The implementation of this ADR will also change the `Data` field of the `Packet` type from `[]byte` (i.e. arbitrary data) to `PacketDataI`. This also removes the `Timeout` field from the `Packet` struct. This is because the `PacketDataI` interface now contains this information. You can see details about this in [ICS04](https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#definitions). 

The `PacketDataI` is the application specific interface that provides information for the execuition of the application packet. In the case of ICS20 this would be `denom`, `amount` and `address`

```go
// PacketDataI defines the standard interface for IBC packet data
type PacketDataI interface {
	GetCommitment() []byte // Commitment form that will be stored in the state.
	GetTimeoutHeight() uint64

	ValidateBasic() sdk.Error
	Type() string
}
```

Example application-side usage:

```go
func NewHandler(k Keeper) sdk.Handler {
  return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
    switch msg := msg.(type) {
    case MsgTransfer:
      return handleMsgTransfer(ctx, k, msg)
    case ibc.MsgPacket:
      switch data := msg.Packet.Data.(type) {
      case PacketDataTransfer: // i.e fulfills the PacketDataI interface
        return handlePacketDataTransfer(ctx, k, msg.Packet, data)
      }
    case ibc.MsgTimeoutPacket:
      switch packet := msg.Packet.Data.(type) {
      case PacketDataTransfer:
        return handleTimeoutPacketDataTransfer(ctx, k, msg.Packet)
      }
    }
  }
}

func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) sdk.Result {
  err := k.SendTransfer(ctx,msg.PortID, msg.ChannelID, msg.Amount, msg.Sender, msg.Receiver)
  if err != nil {
    return sdk.ResultFromError(err)
  }

  return sdk.Result{}
}

func handlePacketDataTransfer(ctx sdk.Context, k Keeper, packet ibc.Packet, data PacketDataTransfer) sdk.Result {
  err := k.ReceiveTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestinationPort(), packet.GetDestinationChannel(), data)
  if err != nil {
    // Source chain sent invalid packet, shutdown channel
  }
  k.PortKeeper.WriteAcknowledgement([]byte{0x00})
  return sdk.Result{}
}

func handleCustomTimeoutPacket(ctx sdk.Context, k Keeper, packet CustomPacket) sdk.Result {
  err := k.RecoverTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestinationPort(), packet.GetDestinationChannel(), data)
  if err != nil {
    // This chain sent invalid packet
    panic(err)
  }
  // packet timeout should not fail
  return sdk.Result{}
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
