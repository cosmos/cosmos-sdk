# ADR 025: IBC Passive Channels

## Changelog

- 2020-05-23: Provide sample Go code and more details
- 2020-05-18: Initial Draft

## Status

Proposed

## Context

The current "naive" IBC Relayer strategy currently establishes a single predetermined IBC channel atop a single connection between two clients (each potentially of a different chain).  This strategy then detects packets to be relayed by watching for `send_packet` and `recv_packet` events matching that channel, and sends the necessary transactions to relay those packets.

We wish to expand this "naive" strategy to a "passive" one which detects and relays both channel handshake messages and packets on a given connection, without the need to know each channel in advance of relaying it.

In order to accomplish this, we propose adding more comprehensive events to expose channel metadata for each transaction sent from the `x/ibc/core/04-channel/keeper/handshake.go` and `x/ibc/core/04-channel/keeper/packet.go` modules.

Here is an example of what would be in `ChanOpenInit`:

```go
const (
  EventTypeChannelMeta = "channel_meta"
  AttributeKeyAction = "action"
  AttributeKeyHops = "hops"
  AttributeKeyOrder = "order"
  AttributeKeySrcPort = "src_port"
  AttributeKeySrcChannel = "src_channel"
  AttributeKeySrcVersion = "src_version"
  AttributeKeyDstPort = "dst_port"
  AttributeKeyDstChannel = "dst_channel"
  AttributeKeyDstVersion = "dst_version"
)
// ...
  // Emit Event with Channel metadata for the relayer to pick up and
  // relay to the other chain
  // This appears immediately before the successful return statement.
  ctx.EventManager().EmitEvents(sdk.Events{
    sdk.NewEvent(
      types.EventTypeChannelMeta,
      sdk.NewAttribute(types.AttributeKeyAction, "open_init"),
      sdk.NewAttribute(types.AttributeKeySrcConnection, connectionHops[0]),
      sdk.NewAttribute(types.AttributeKeyHops, strings.Join(connectionHops, ",")),
      sdk.NewAttribute(types.AttributeKeyOrder, order.String()),
      sdk.NewAttribute(types.AttributeKeySrcPort, portID),
      sdk.NewAttribute(types.AttributeKeySrcChannel, chanenlID),
      sdk.NewAttribute(types.AttributeKeySrcVersion, version),
      sdk.NewAttribute(types.AttributeKeyDstPort, counterparty.GetPortID()),
      sdk.NewAttribute(types.AttributeKeyDstChannel, counterparty.GetChannelID()),
      // The destination version is not yet known, but a value is necessary to pad
      // the event attribute offsets
      sdk.NewAttribute(types.AttributeKeyDstVersion, ""),
    ),
  })
```

These metadata events capture all the "header" information needed to route IBC channel handshake transactions without requiring the client to query any data except that of the connection ID that it is willing to relay.  It is intended that `channel_meta.src_connection` is the only event key that needs to be indexed for a passive relayer to function.

### Handling Channel Open Attempts

In the case of the passive relayer, when one chain sends a `ChanOpenInit`, the relayer should inform the other chain of this open attempt and allow that chain to decide how (and if) it continues the handshake.  Once both chains have actively approved the channel opening, then the rest of the handshake can happen as it does with the current "naive" relayer.

To implement this behavior, we propose replacing the `cbs.OnChanOpenTry` callback with a new `cbs.OnAttemptChanOpenTry` callback which explicitly handles the `MsgChannelOpenTry`, usually by resulting in a call to `keeper.ChanOpenTry`.  The typical implementation, in `x/ibc-transfer/module.go` would be compatible with the current "naive" relayer, as follows:

```go
func (am AppModule) OnAttemptChanOpenTry(
  ctx sdk.Context,
  chanKeeper channel.Keeper,
  portCap *capability.Capability,
  msg channel.MsgChannelOpenTry,
) (*sdk.Result, error) {
  // Require portID is the portID transfer module is bound to
  boundPort := am.keeper.GetPort(ctx)
  if boundPort != msg.PortID {
    return nil, sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", msg.PortID, boundPort)
  }

  // BEGIN NEW CODE
  // Assert our protocol version, overriding the relayer's suggestion.
  msg.Version = types.Version
  // Continue the ChanOpenTry.
  res, chanCap, err := channel.HandleMsgChannelOpenTry(ctx, chanKeeper, portCap, msg)
  if err != nil {
    return nil, err
  }
  // END OF NEW CODE

  // ... the rest of the callback is similar to the existing OnChanOpenTry
  // but uses msg.* directly.
```

Here is how this callback would be used, in the implementation of `x/ibc/handler.go`:

```go
// ...
    case channel.MsgChannelOpenTry:
      // Lookup module by port capability
      module, portCap, err := k.PortKeeper.LookupModuleByPort(ctx, msg.PortID)
      if err != nil {
        return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
      }
      // Retrieve callbacks from router
      cbs, ok := k.Router.GetRoute(module)
      if !ok {
        return nil, sdkerrors.Wrapf(port.ErrInvalidRoute, "route not found to module: %s", module)
      }
      // Delegate to the module's OnAttemptChanOpenTry.
      return cbs.OnAttemptChanOpenTry(ctx, k.ChannelKeeper, portCap, msg)
```

The reason we do not have a more structured interaction between `x/ibc/handler.go` and the port's module (to explicitly negotiate versions, etc) is that we do not wish to constrain the app module to have to finish handling the `MsgChannelOpenTry` during this transaction or even this block.

## Decision

- Expose events to allow "passive" connection relayers.
- Enable application-initiated channels via such passive relayers.
- Allow port modules to control how to handle open-try messages.

## Consequences

### Positive

Makes channels into a complete application-level abstraction.

Applications have full control over initiating and accepting channels, rather than expecting a relayer to tell them when to do so.

A passive relayer does not have to know what kind of channel (version string, ordering constraints, firewalling logic) the application supports.  These are negotiated directly between applications.

### Negative

Increased event size for IBC messages.

### Neutral

More IBC events are exposed.

## References

- The Agoric VM's IBC handler currently [accomodates `attemptChanOpenTry`](https://github.com/Agoric/agoric-sdk/blob/904b3a0423222a1b32893453e44bbde598473960/packages/cosmic-swingset/lib/ag-solo/vats/ibc.js#L546)
