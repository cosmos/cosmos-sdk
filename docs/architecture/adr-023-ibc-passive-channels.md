# ADR 023: IBC Passive Channels

## Changelog

- 2020-05-18: Initial Draft

## Context

The current "naive" IBC Relayer strategy currently establishes a single predetermined IBC channel atop a single connection between two clients (each potentially of a different chain).  This strategy then detects packets to be relayed by watching for `send_packet` and `recv_packet` events matching that channel, and sends the necessary transactions to relay those packets.

We wish to expand this "naive" strategy to a "passive" one which detects and relays both channel handshake messages and packets on a given connection, without the need to know each channel in advance of relaying it.

In order to accomplish this, we propose adding the following events to expose channel metadata for each transaction sent from the `x/ibc/04-channel/keeper/handshake.go` and `x/ibc/04-channel/keeper/packet.go` modules:

- `channel_meta.src_connection=CONN1` as the only key needing to be indexed
- `channel_meta.action=ACTION` where `ACTION` is one of:
  - `open_init`
  - `open_try`
  - `open_ack`
  - `open_confirm`
  - `send_packet`
  - `packet_executed`
  - `close_init`
  - `close_confirm`
- `channel_meta.hops=CONN1,CONN2,...`
- `channel_meta.order=ORDERED`
- `channel_meta.src_port=PORT1`
- `channel_meta.src_channel=CHANNEL1`
- `channel_meta.src_version=VSN1`
- `channel_meta.dst_port=PORT2`
- `channel_meta.dst_channel=CHANNEL2`
- `channel_meta.dst_version=VSN2`

These metadata events capture all the "header" information needed to route IBC channel handshake transactions without requiring the client to keep track of any state except for its connection ID.

### Inversion of Control

The other concern we are trying to address is IBC use-cases where an application module wants to fully control the opening and closing of channels, both initiating these requests and deciding how to proceed with them.

Initiation is straightforward: just emit the above events to allow a relayer to notice them.  Handling of requests needs a different architecture, as the current IBC implementation presumes that the relayer is in control of setting up and tearing down each connection and can unilaterally impose its will on the chain's IBC stack.  The IBC Channel messages are handled directly by the IBC implementation and not reflected to the application until after they have been processed.

We propose that as an alternative to this behaviour, an IBC application module could opt-in to marking a routed IBC port as "controlled".  This flag would prevent the IBC handler from directly calling the various IBC keepers (except for behaviour already defined by the `send_packet` and `recv_packet` events).  Instead "channel handshake" messages provided by the relayer would inform the application of the other side's intentions, allowing the application to decide how and when to call the IBC keepers to continue the handshake.

The `.OnChanHandshake` callback would receive a wrapped `channel.MsgChannelHandshake{}` message:

```go
typedef MsgChannelHandshake struct {
  // Initiator is the IBC channel message that the relayer noticed.
  // This can be any of the MsgChan* messages, populated by the data from
  // the channel_meta events.
  Initiator   sdk.Msg
  // Proof from the sender's ibctypes.KeyChannel(SrcPort, SrcChannel)
  Proof       commitmentexported.Proof,
  ProofHeight uint64
  // The relayer's signature.
  Signer      sdk.AccAddress
}
```

## Decision

- Expose events to allow "passive" connection relayers.
- Enable application-initiated channels via such passive relayers.
- Allow ports to opt-in to explicit "channel handshake" messages so they can control their fate.

## Status

Proposed

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## References

- {reference link}
