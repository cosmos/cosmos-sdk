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

It is strongly recommended to read the full details of [ADR 001: Coin Source Tracing](./../../../../../docs/architecture/adr-001-coin-source-tracing.md) to understand the implications and context of the IBC token representations.

### UX suggestion for clients

For clients that want to display the source of the token, it is recommended to use the following
alternatives for each of the following cases:

#### Direct connection

If the denomination trace contains a single identifier prefix pair (as in the example above), then
the easiest way to retrieve the chain and light client identifier is to map the trace information
directly. In summary, this requires querying the channel from the denomination trace identifiers,
and then the counterparty client state using the counterparty port and channel identifiers from the
retrieved channel.

A general pseudo algorithm would look like the following:

1. Query the full denomination trace.
2. Query the channel with the `portID/channelID` pair, which corresponds to the first destination of the
   token.
3. Query the client state using the identifiers pair. Note that this query will return a `"Not
   Found"` response if the current chain is not connected to this channel.
4. Retrieve the the client identifier or chain identifier from the client state (eg: on
   Tendermint clients) and store it locally.

Using the gRPC gataway client service the steps above would be, with a given IBC token `ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2`:

1. `GET /ibc_transfer/v1beta1/denom_traces/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2` -> `{"path": "transfer/channelToA", "base_denom": "uatom"}`
2. `GET /ibc/channel/v1beta1/channels/channelToA/ports/transfer/client_state"` -> `{"client_id": "clientA", "chain-id": "chainA", ...}`
3. `GET /ibc/channel/v1beta1/channels/channelToA/ports/transfer"` -> `{"channel_id": "channelToA", port_id": "transfer", counterparty: {"channel_id": "channelToB", port_id": "transfer"}, ...}`
4. `GET /ibc/channel/v1beta1/channels/channelToB/ports/transfer/client_state" -> {"client_id": "clientB", "chain-id": "chainB",}`

Then, the token transfer chain path for the `uatom` denomination would be: `chainB` -> `chainA`.

### Multiple connection hops

The IBC protocol doesn't know the topology of the overall network (i.e connections between chains and identifier names between them). In the concrete case of fungible token transfers with more than one connection hop, a particular chain in the timeline of the individual transfers can't query the chain and client identifiers of the other chains.

Take for example the following sequence of transfers `A -> B -> C` for an IBC token, with a final prefix path (trace info) of `transfer/channelChainC/transfer/channelChainB`. What the paragraph above means is that is that even in the case that chain `C` is directly connected to chain `A`, querying the port and channel identifiers that chain `B` uses to connect to chain `A` (eg: `transfer/channelChainA`) can be completely different from the one that chain `C` uses to connect to chain `A` (eg: `transfer/channelToChainA`).

Thus the proposed solution for clients that the IBC team recommends are the following:

- **Connect to all chains**: Connecting to all the chains in the timeline would allow clients to
  perform the queries outlined in the [direct connection](#direct-connection) section to each
  relevant chain. By repeatedly following the port and channel denomination trace transfer timeline,
  clients should always be able to find all the relevant identifiers. This comes at the tradeoff
  that the client must connect to nodes on each of the chains in order to perform the queries.
- **Relayer as a Service (RaaS)**: A longer term solution is to use/create a relayer service that
  could map the denomination trace to the chain path timeline for each token (i.e `origin chain ->
  chain #1 -> ... -> chain #(n-1) -> final chain`). Clients would be advised to connect to public
  relayers that support the largest number of connections between chains in the ecosystem. Unfortunately, none of the existing public relayers (in [Golang](https://github.com/cosmos/relayer) and [Rust](https://github.com/informalsystems/ibc-rs)), provide this service to clients.

::: tip
The only viable alternative for clients (at the time of writing) to tokens with multiple connection hops, is to connect to all chains directly and perform relevant queries to each of them in the sequence.
:::

## Locked Funds

In some [exceptional cases](./../../../../../docs/architecture/adr-026-ibc-client-recovery-mechanisms.md#exceptional-cases), a client state associated with a given channel cannot be updated. This causes that funds from fungible tokens in that channel will be permanently locked and thus can no longer be transferred.

To mitigate this, a client update governance proposal can be submitted to update the frozen client
with a new valid header. Once the proposal passes the client state will be unfrozen and the funds
from the associated channels will then be unlocked. This mechanism only applies to clients that
allow updates via governance, such as Tendermint clients.
