<!--
order: 1
-->

# Concepts

## Acknowledgements

ICS20 uses the recommended acknowledgement format as specified by [ICS 04](https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#acknowledgement-envelope).

A successful receive of a transfer packet will result in a Result Acknowledgement being written
with the value `[]byte(byte(1))` in the `Response` field.

An unsuccessful receive of a transfer packet will result in an Error Acknowledgement being written
with the error message in the `Response` field.

## Denomination Trace

The denomination trace corresponds to the information that allows a token to be traced back to its
origin chain. It contains a sequence of port and channel identifiers ordered from the most recent to
the oldest in the timeline of transfers.

This information is included on the token denomination field in the form of a hash to prevent an
unbounded denomination length. For example, the token `transfer/channelToA/uatom` will be displayed
as `ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2`.

Each send to any chain other than the one it was previously received from is a movement forwards in
the token's timeline. This causes trace to be added to the token's history and the destination port
and destination channel to be prefixed to the denomination. In these instances the sender chain is
acting as the "source zone". When the token is sent back to the chain it previously received from, the
prefix is removed. This is a backwards movement in the token's timeline and the sender chain is
acting as the "sink zone".

### UX suggestion for clients

For clients that want to display the source of the token, it is recommended to use one of the following alternatives:

- **Use a relayer service**: Available relayer services could map the denomination trace to the
  chain path timeline for each token (i.e `origin chain -> chain #1 -> ... -> chain #(n-1) -> final
  chain`). Clients are advised to connect to public relayers that support the largest number of
  connections between chains in the ecosystem.
- **Map trace information directly:** If the chain you are connected is also connected to each of
  the channels from the denomination trace information, then it is theoretically possible to
  retrieve the chain path timeline directly through a series of queries:

::: tip
 From a chain's perspective, the IBC protocol doesn't know the topology of the network (i.e
 connections between chains and identifier names between them). The proposed solution only works if
 the identifiers have consistent names between the chains. Because of this, it is strongly
 recommended to use a relayer service instead.
:::

1. Query the full denomination trace.
1. For each port and channel identifier pair (from left to right) in the trace info:
    1. Query the client state using the identifiers pair. Note that this query will return a `"Not
       Found"` response if the current chain is not connected to this channel.
    1. Retrieve the the client identifier or chain identifier from the client state and (eg: on
       Tendermint clients) insert it to a map or array as desired.
1. Query the channel with the right-most pair, which corresponds to the first destination of the
   token.
1. Query the the client state (like on step 2.1) using the counterparty port and channel identifiers
   from the above result.
1. Retrieve the the client identifier or chain identifier like on 2.2.

#### Example

Using the gRPC gataway client service the steps above would be, with a given IBC token `ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2`:

1. `GET /ibc_transfer/v1beta1/denom_traces/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2` -> `{"path": "transfer/channelToA", "base_denom": "uatom"}`
        1. `GET /ibc/channel/v1beta1/channels/channelToA/ports/transfer/client_state"` -> `{"client_id": "clientA", "chain-id": "chainA", ...}`
1. `GET /ibc/channel/v1beta1/channels/channelToA/ports/transfer"` -> `{"channel_id": "channelToA", port_id": "transfer", counterparty: {"channel_id": "channelToB", port_id": "transfer"}, ...}`
1. `GET /ibc/channel/v1beta1/channels/channelToB/ports/transfer/client_state" -> {"client_id": "clientB", "chain-id": "chainB",}

Then, the token transfer chain path for the `uatom` denomination would be: `chainB` -> `chainA`.

## Locked Funds

In some [exceptional cases](./../../../../../docs/architecture/adr-026-ibc-client-recovery-mechanisms.md#exceptional-cases), a client state associated with a given channel cannot be updated. This causes that funds from fungible tokens in that channel will be permanently locked and thus can no longer be transferred.

To mitigate this, a client update governance proposal can be submitted to update the frozen client
with a new valid header. Once the proposal passes the client state will be unfrozen and the funds
from the associated channels will then be unlocked. This mechanism only applies to clients that
allow updates via governance, such as Tendermint clients.
