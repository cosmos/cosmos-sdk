<!--
order: 3
-->

# Msg Service

## Msg/OptIn

An account can activate quarantine using a `MsgOptIn`.
It contains only the address to quarantine.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/tx.proto#L33-L38

It is expected to fail if the `to_address` is invalid.

## Msg/OptOut

An account can deactivate quarantine using a `MsgOptOut`.
It contains only the address to unquarantine.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/tx.proto#L43-L48

It is expected to fail if the `to_address` is invalid.

## Msg/Accept

Quarantined funds can be accepted by the intended receiver using a `MsgAccept`.
It contains a `to_address` (receiver) and one or more `from_addresses` (senders).
It also contains a flag to indicate whether auto-accept should be set up for all provided addresses.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/tx.proto#L53-L67

Any quarantined funds for the `to_address` from any `from_address` are accepted (regardless of whether they've been previously declined).

For quarantined funds from multiple senders (e.g. from a `MultiSend`), all senders must be part of an `Accept` before the funds will be released,
but they don't all have to be part of the same `Accept`.

If the `permanent` flag is `true`, after accepting all applicable funds, auto-accept is set up to the `to_address` from each of the provided `from_addresses`.

It is expected to fail if:
- The `to_address` is missing or invalid.
- No `from_addresses` are provided.
- Any `from_addresses` are invalid.

The response will contain a total of all funds released.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/tx.proto#L69-L74

## Msg/Decline

Quarantined funds can be declined by the intended receiver using a `MsgDecline`.
It contains a `to_address` (receiver) and one or more `from_addresses` (senders).
It also contains a flag to indicate whether auto-decline should be set up for all provided addresses.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/tx.proto#L76-L90

Any quarantined funds for the `to_address` from any `from_address` are declined.

For quarantined funds from multiple senders (e.g. from a `MultiSend`), a decline from any sender involved is sufficient to decline the funds.
Funds that have been declined can always be accepted later.

If the `permanent` flag is `true`, after declining all applicable funds, auto-decline is set up to the `to_address` from each of the provided `from_addresses`.

It is expected to fail if:
- The `to_address` is missing or invalid.
- No `from_addresses` are provided.
- Any `from_addresses` are invalid.

## Msg/UpdateAutoResponses

Auto-Responses can be defined either through the `permanent` flags with a `MsgAccept` or `MsgDecline`, or using a `MsgUpdateAutoResponses`.
It contains a `to_address` and a list of `updates`. Each `AutoResponseUpdate` contains a `from_address` and the desired `response` for it.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/tx.proto#L95-L104

Providing a `response` of `AUTO_RESPONSE_UNSPECIFIED` will cause the applicable entry to be deleted, allowing users to un-set previous auto-responses.

Updating auto-responses has no effect on existing quarantined funds.

It is expected to fail if:
- The `to_address` is invalid.
- No `updates` are provided. 
- Any `from_address` is missing or invalid.
- Any `response` value is something other than `AUTO_RESPONSE_ACCEPT`, `AUTO_RESPONSE_DECLINE`, or `AUTO_RESPONSE_UNSPECIFIED`.  