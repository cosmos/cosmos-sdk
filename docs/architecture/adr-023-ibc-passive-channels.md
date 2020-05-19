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

### Accepting Channel Opens

The other concern we are trying to address is IBC use-cases where an application module wants to fully control the opening of channels, by being able to inspect connection init metadata and decide how to complete the handshake.

The `handler.OnChanOpenTry` handler would call the `cbs.OnChanOpenAccept` callback with the same arguments as `OnChanOpenTry`.  If this callback is not specified, the default behaviour would be the current one: to call the corresponding `keeper.ChanOpenTry` with the supplied values.  If the `.OnChanOpenAccept` is specified, then it would have the ability to call the `keeper.ChanOpenTry` how and when it would like (if at all).

## Decision

- Expose events to allow "passive" connection relayers.
- Enable application-initiated channels via such passive relayers.
- Allow ports to control which channel open attempts they honour.

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
