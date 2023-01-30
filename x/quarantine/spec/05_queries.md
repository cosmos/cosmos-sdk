<!--
order: 5
-->

# gRPC Queries

## Query/IsQuarantined

To find out if an account is quarantined, use `QueryIsQuarantinedRequest`.
The query takes in a `to_address` and outputs `true` if the address is quarantined, or `false` otherwise.

Request:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/query.proto#L46-L50

Response:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/query.proto#L52-L56

It is expected to fail if the `to_address` is invalid.

## Query/QuarantinedFunds

To get information on quarantined funds, use `QueryQuarantinedFundsRequest`.
This query takes in an optional `to_address` and optional `from_address` and outputs information on quarantined funds.

Request:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/query.proto#L58-L67

Response:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/query.proto#L69-L76

QuarantinedFunds:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/quarantine.proto#L10-L21

- If neither a `to_address` nor `from_address` are provided, all non-declined quarantined funds for any addresses will be returned.
- If the request contains a `to_address` but no `from_address`, all non-declined quarantined funds for the `to_address` are returned.
- If both a `to_address` and `from_address` are provided, all quarantined funds to the `to_address` involving the `from_address` a returned regardless of whether they've been declined.

This query is paginated.

It is expected to fail if:
- A `from_address` is provided without a `to_address`.
- Either the `to_address` or `from_address` is provided but invalid.
- Invalid pagination parameters are provided.

## Query/AutoResponses

To see the auto-response settings, use `QueryAutoResponsesRequest`.
This query takes in a `to_address` and optional `from_address` and outputs information about auto-responses.

Request:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/query.proto#L78-L87

Response:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/query.proto#L89-L96

AutoResponseEntry:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/quarantine/v1beta1/quarantine.proto#L23-L31

- If no `from_address` is provided, all auto-response entries for the provided `to_address` are returned. The results will not contain any entries for `AUTO_RESPONSE_UNSPECIFIED`.
- If a `from_address` is provided, the auto-response setting that `to_address` has from `from_address` is returned. This result might be `AUTO_RESPONSE_UNSPECIFIED`.

This query is paginated.

It is expected to fail if:
- The `to_address` is empty or invalid.
- A `from_address` is provided and invalid.
- Invalid pagination parameters are provided.

