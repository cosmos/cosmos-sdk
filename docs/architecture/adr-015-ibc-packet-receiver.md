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
  var flag bool

  for _, msg := range tx.GetMsgs() {
    var err error
    switch msg := msg.(type) {
    case client.MsgUpdateClient:
      err = pvr.clientKeeper.UpdateClient(msg.ClientID, msg.Header)
    case channel.MsgPacket:
      err = pvr.channelKeeper.VerifyPacket(msg.Packet, msg.Proofs, msg.ProofHeight)
      flag = true
      // Store the empty acknowledgement for convinience
      pvr.channelKeeper.SetPacketAcknowledgement(ctx, msg.PortID, msg.ChannelID, msg.Sequence, []byte{})
    case chanel.MsgAcknowledgement:
      err = pvr.channelKeeper.VerifyAcknowledgement(msg.Acknowledgement, msg.Proof, msg.ProofHeight)
      flag = true
    case channel.MsgTimeoutPacket:
      err = pvr.channelKeeper.VerifyTimeout(msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv)
      flag = true
    default:
      if flag {
        return ctx, errors.New("Transaction cannot include both IBC packet messasges and normal messages")
      }
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

An attacker can insert a failiing message before any of packet receiving message
preventing the packet messages to be processed but keeping the sequence increased.
If a transaction contains a packet receiving messages, any possibly failing 
messages should not be included in the transaction.

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
    case ibc.MsgPacket:
      return k.port.WriteAcknowledgement(ctx, msg, func(ctx sdk.Context, p ibc.PacketDataI) sdk.Result {
        switch packet := p.(type) {
        case CustomPacket: // i.e fulfills the PacketDataI interface
          return handleCustomPacket(ctx, k, packet)
        }
      })
    case ibc.MsgAcknowledgement:
      switch ack := msg.Acknowledgement.Data.(type) {
      case CustomAcknowledgement:
        return handleCustomAcknowledgement(ctx, k, msg.Acknowledgement)
      }
    case ibc.MsgTimeoutPacket:
      switch packet := msg.Packet.Data.(type) {
      case CustomPacket:
        return handleCustomTimeoutPacket(ctx, k, msg.Packet)
      }
    }
  }
}

func handleCustomPacket(ctx sdk.Context, k Keeper, packet CustomPacket) sdk.Result {
  if failureCondition {
    return AckInvalidPacketContent(k.codespace, []byte{packet.Error()})
  }
  // Handler logic
  return sdk.Result{}
}

func handleCustomAcknowledgement(ctx sdk.Context, k Keeper, ack MyAcknowledgement) sdk.Result {
  // Handler logic
  return sdk.Result{}
}

func handleCustomTimeoutPacket(ctx sdk.Context, k Keeper, packet CustomPacket) sdk.Result {
  // Handler logic
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
