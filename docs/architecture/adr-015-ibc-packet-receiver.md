# ADR 015: IBC Packet Receiver

## Changelog

- 2019 Oct 22: Initial Draft

## Context
 
[ICS 26 - Routing Module](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module) defines a function [`handlePacketRecv`](https://github.com/cosmos/ics/tree/master/spec/ics-026-routing-module#packet-relay).

In ICS 26, the routing module is defined as a layer above each application module
which verifies and routes messages to the destination modules. It is possible to
implement it as a separate module, however, we already have functionality to route
messages upon the destination identifiers in the baseapp. This ADR suggests
to utilize existing `baseapp.router` to route packets to application modules.

Generally, routing module callbacks have two separate steps in them,
verification and execution. This corresponds to the `AnteHandler`-`Handler`
model inside the SDK. We can do the verification inside the `AnteHandler`
in order to increase developer ergonomics by reducing boilerplate
verification code.

For atomic multi-message transaction, we want to keep the IBC related
state modification to be preserved even the application side state change
reverts. One of the example might be IBC token sending message following with
stake delegation which uses the tokens received by the previous packet message.
If the token receiving fails for any reason, we might not want to keep
executing the transaction, but we also don't want to abort the transaction
or the sequence and commitment will be reverted and the channel will be stuck.
This ADR suggests new `CodeType`, `CodeTxBreak`, to fix this problem.

## Decision

`PortKeeper` will have the capability key that is able to access only the
channels bound to the port. Entities that hold a `PortKeeper` will be
able to call the methods on it which are corresponding with the methods with
the same names on the `ChannelKeeper`, but only with the
allowed port. `ChannelKeeper.Port(string, ChannelChecker)` will be defined to
easily construct a capability-safe `PortKeeper`. This will be addressed in
another ADR and we will use insecure `ChannelKeeper` for now.

`baseapp.runMsgs` will break the loop over the messages if one of the handlers
returns `!Result.IsOK()`. However, the outer logic will write the cached
store if `Result.IsOK() || Result.Code.IsBreak()`. `Result.Code.IsBreak()` if
`Result.Code == CodeTxBreak`.

```go
func (app *BaseApp) runTx(tx Tx) (result Result) {
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
  // BEGIN modification made in this ADR
  if result.IsOK() || result.IsBreak() {
  // END
    msCache.Write()
  }

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

func (pvr ProofVerificationDecorator) AnteHandle(ctx Context, tx Tx, simulate bool, next AnteHandler) (Context, error) {
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
    case channel.MsgChannelOpenInit;
      err = pvr.channelKeeper.CheckOpen(msg.PortID, msg.ChannelID, msg.Channel)
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
`VerifyTimeout` will be extracted out into separated functions,
`WriteAcknowledgement`, `DeleteCommitment`, `DeleteCommitmentTimeout`, respectively,
which will be called by the application handlers after the execution.

`WriteAcknowledgement` writes the acknowledgement to the state that can be
verified by the counter-party chain and increments the sequence to prevent
double execution. `DeleteCommitment` will delete the commitment stored,
`DeleteCommitmentTimeout` will delete the commitment and close channel in case
of ordered channel.

```go
func (keeper ChannelKeeper) WriteAcknowledgement(ctx Context, packet Packet, ack []byte) {
  keeper.SetPacketAcknowledgement(ctx, packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(), ack)
  keeper.SetNextSequenceRecv(ctx, packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
}

func (keeper ChannelKeeper) DeleteCommitment(ctx Context, packet Packet) {
  keeper.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
}

func (keeper ChannelKeeper) DeleteCommitmentTimeout(ctx Context, packet Packet) {
  k.deletePacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
  
  if channel.Ordering == types.ORDERED [
    channel.State = types.CLOSED
    k.SetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), channel)
  }
}
```

Each application handler should call respective finalization methods on the `PortKeeper`
in order to increase sequence (in case of packet) or remove the commitment
(in case of acknowledgement and timeout).
Calling those functions implies that the application logic has successfully executed. 
However, the handlers can return `Result` with `CodeTxBreak` after calling those methods
which will persist the state changes that has been already done but prevent any further 
messages to be executed in case of semantically invalid packet. This will keep the sequence
increased in the previous IBC packets(thus preventing double execution) without 
proceeding to the following messages.
In any case the application modules should never return state reverting result, 
which will make the channel unable to proceed.

`ChannelKeeper.CheckOpen` method will be introduced. This will replace `onChanOpen*` defined 
under the routing module specification. Instead of define each channel handshake callback
functions, application modules can provide `ChannelChecker` function with the `AppModule`
which will be injected to `ChannelKeeper.Port()` at the top level application.
`CheckOpen` will find the correct `ChennelChecker` using the
`PortID` and call it, which will return an error if it is unacceptable by the application.

The `ProofVerificationDecorator` will be inserted to the top level application.
It is not safe to make each module responsible to call proof verification
logic, whereas application can misbehave(in terms of IBC protocol) by
mistake.

The `ProofVerificationDecorator` should come right after the default sybil attack
resistent layer from the current `auth.NewAnteHandler`:

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

The implementation of this ADR will also create a `Data` field of the `Packet` of type `[]byte`, which can be deserialised by the receiving module into its own private type. It is up to the application modules to do this according to their own interpretation, not by the IBC keeper.  This is crucial for dynamic IBC.

Example application-side usage:

```go
type AppModule struct {}

// CheckChannel will be provided to the ChannelKeeper as ChannelKeeper.Port(module.CheckChannel)
func (module AppModule) CheckChannel(portID, channelID string, channel Channel) error {
  if channel.Ordering != UNORDERED {
    return ErrUncompatibleOrdering()
  }
  if channel.CounterpartyPort != "bank" {
    return ErrUncompatiblePort()
  }
  if channel.Version != "" {
    return ErrUncompatibleVersion()
  }
  return nil
}

func NewHandler(k Keeper) Handler {
  return func(ctx Context, msg Msg) Result {
    switch msg := msg.(type) {
    case MsgTransfer:
      return handleMsgTransfer(ctx, k, msg)
    case ibc.MsgPacket:
      var data PacketDataTransfer
      if err := types.ModuleCodec.UnmarshalBinaryBare(msg.GetData(), &data); err != nil {
        return err
      }
      return handlePacketDataTransfer(ctx, k, msg, data)
    case ibc.MsgTimeoutPacket:
      var data PacketDataTransfer
      if err := types.ModuleCodec.UnmarshalBinaryBare(msg.GetData(), &data); err != nil {
        return err
      }
      return handleTimeoutPacketDataTransfer(ctx, k, packet)
    // interface { PortID() string; ChannelID() string; Channel() ibc.Channel }
    // MsgChanInit, MsgChanTry implements ibc.MsgChannelOpen
    case ibc.MsgChannelOpen: 
      return handleMsgChannelOpen(ctx, k, msg)
    }
  }
}

func handleMsgTransfer(ctx Context, k Keeper, msg MsgTransfer) Result {
  err := k.SendTransfer(ctx,msg.PortID, msg.ChannelID, msg.Amount, msg.Sender, msg.Receiver)
  if err != nil {
    return sdk.ResultFromError(err)
  }

  return sdk.Result{}
}

func handlePacketDataTransfer(ctx Context, k Keeper, packet Packet, data PacketDataTransfer) Result {
  err := k.ReceiveTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestinationPort(), packet.GetDestinationChannel(), data)
  if err != nil {
    // TODO: Source chain sent invalid packet, shutdown channel
  }
  k.ChannelKeeper.WriteAcknowledgement([]byte{0x00}) // WriteAcknowledgement increases the sequence, preventing double spending
  return sdk.Result{}
}

func handleCustomTimeoutPacket(ctx Context, k Keeper, packet CustomPacket) Result {
  err := k.RecoverTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestinationPort(), packet.GetDestinationChannel(), data)
  if err != nil {
    // This chain sent invalid packet or cannot recover the funds
    panic(err)
  }
  k.ChannelKeeper.DeleteCommitmentTimeout(ctx, packet)
  // packet timeout should not fail
  return sdk.Result{}
}

func handleMsgChannelOpen(sdk.Context, k Keeper, msg MsgOpenChannel) Result {
  k.AllocateEscrowAddress(ctx, msg.ChannelID())
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

### Neutral

- Introduces new `AnteHandler` decorator.
- Dynamic ports can be supported using hierarchical port identifier, see #5290 for detail

## References

- Relevant comment: [cosmos/ics#289](https://github.com/cosmos/ics/issues/289#issuecomment-544533583)
- [ICS26 - Routing Module](https://github.com/cosmos/ics/blob/master/spec/ics-026-routing-module)
