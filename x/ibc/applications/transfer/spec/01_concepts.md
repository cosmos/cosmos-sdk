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

This information is included on the token denomination field in the form of a hash to prevent an unbounded denomination length. For example, the token `transfer/channelToA/uatom` will be displayed as
`ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2`.

Each send to any chain other than the one it was previously received from is a movement forwards in
the token's timeline. This causes trace to be added to the token's history and the destination port
and destination channel to be prefixed to the denomination. In these instances the sender chain is
acting as the "source zone". When the token is sent back to the chain it previously received from, the
prefix is removed. This is a backwards movement in the token's timeline and the sender chain is
acting as the "sink zone".

For clients that want to display the source of the token, it is recommended to:

1. Query the full denomination trace.
2. Query the channel using the 2 left most identifiers from the trace. These correspond the first destination port and channel identifiers.
3. Query the client state using the channel's counterparty port and channel identifiers.

The client state can then contain useful information such as the client identifier or chain identifier (eg: on Tendermint clients).

## Send Fungible Tokens

## Receive Fungible Tokens

## Locked Funds
