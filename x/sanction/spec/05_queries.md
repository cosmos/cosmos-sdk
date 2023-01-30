<!--
order: 5
-->

# gRPC Queries

## Query/IsSanctioned

To find out if an account is sanctioned, use `QueryIsSanctionedRequest`.
The query takes in an `address` and outputs whether the account `is_sanctioned`.

This query takes into account any temporary sanctions or unsanctions.
If it returns `true`, the account is not allowed to move its funds.
If it returns `false`, the account *is* allowed to move its funds (at least from a sanction perspective).

Request:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/query.proto#L34-L37

Response:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/query.proto#L39-L43

It is expected to fail if the `address` is invalid.

## Query/SanctionedAddresses

To get all addresses that have permanent sanctions, use `QuerySanctionedAddressesRequest`.
It takes in `pagination` parameters and outputs a list of `addresses`.

Request:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/query.proto#L45-L49

Response:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/query.proto#L51-L58

This query does not take into account temporary sanctions or temporary unsanctions. 
Addresses that are temporarily sanctioned (but not permanently sanctioned) are **not** returned by this query.
Addresses that are permanently sanctioned but temporarily unsanctioned **are** returned by this query.

This query is paginated.

It is expected to fail if invalid `pagination` parameters are provided.

## Query/TemporaryEntries

To get information about temporarily sanctioned or unsanctioned accounts, use `QueryTemporaryEntriesRequest`.
It takes in `pagination` parameters and an optional `address`.

Request:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/query.proto#L60-L67

Response:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/query.proto#L69-L75

TemporaryEntry:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/sanction.proto#L27-L35

TempStatus:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/sanction.proto#L37-L47

- If an `address` is provided, only temporary entries associated with that address are returned.
- If an `address` is provided that does not have any temporary entries, a single `TemporaryEntry` with a `status` of `TEMP_STATUS_UNSPECIFIED` is returned.
  Otherwise only entries with a `status` of `TEMP_STATUS_SANCTIONED` or `TEMP_STATUS_UNSANCTIONED` are returned.
- If an `address` is not provided, all temporary entries are returned.

This query is paginated.

It is expected to fail if:
- An `address` is provided that is invalid.
- Invalid `pagination` parameters are provided.

## Query/Params

To get the `x/sanction` module's params, use `QueryParamsRequest`.
It has no input and outputs the `params`.

Request:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/query.proto#L77-L78

Response:

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/query.proto#L80-L84

This query returns the values used for the params.
That is, if there are params stored in state, they are returned;
if there aren't params stored in state, the default values are returned.

It is not expected to fail.