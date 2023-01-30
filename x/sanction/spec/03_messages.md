<!--
order: 3
-->

# Msg Service

All Msg Service endpoints in the `x/sanction` module are for use with governance proposals.

## Msg/Sanction

A user can request that accounts be sanctioned by submitting a governance proposal containing a `MsgSanction`.
It contains the list of `addresses` of accounts to be sanctioned and the `authority` able to do it.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/tx.proto#L22-L32

If the proposal ever has enough total deposit (defined in params), immediate temporary sanctions are issued for each address.
Temporary sanctions expire at the completion of the governance proposal regardless of outcome.

If the proposal passes, permanent sanctions are enacted for each address and temporary entries for each address are removed.
Otherwise, any temporary entries associated with the governance proposal are removed.

It is expected to fail if:
- The `authority` provided does not equal the authority defined for the `x/sanction` module's keeper.
  This is most often the address of the `x/gov` module's account.
- Any `addresses` are not valid bech32 encoded address strings.
- Any `addresses` are unsanctionable.

## Msg/Unsanction

A user can request that accounts be unsanctioned by submitting a governance proposal containing a `MsgUnsanction`.
It contains the list of `addresses` of accounts to be unsanctioned and the `authority` able to do it.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/tx.proto#L37-L47

If the proposal ever has enough total deposit (defined in params), immediate temporary unsanctions are issued for each address.
Temporary unsanctions expire at the completion of the governance proposal regardless of outcome.

If the proposal passes, permanent sanctions are removed for each address and temporary entries for each address are also removed.
Otherwise, any temporary entries associated with the governance proposal are removed.

It is expected to fail if:
- The `authority` provided does not equal the authority defined for the `x/sanction` module's keeper.
  This is most often the address of the `x/gov` module's account.
- Any `addresses` are not valid bech32 encoded address strings.

## Msg/UpdateParams

The sanction module params can be updated by submitting a governance proposal containing a `MsgUpdateParams`.
It contains the desired new `params` and the `authority` able to update them.

+++ https://github.com/provenance-io/cosmos-sdk/blob/da2ea8a8139ae9e110de0776baffa1d0dd97db5e/proto/cosmos/sanction/v1beta1/tx.proto#L52-L62

If `params` is `null`, they will be deleted from state, reverting them to their code-defined defaults.
If a field in `params` is `null` or empty, the record in state will reflect that.

It is expected to fail if:
- The `authority` provided does not equal the authority defined for the `x/sanction` module's keeper.
  This is most often the address of the `x/gov` module's account.
- Any params are invalid.
