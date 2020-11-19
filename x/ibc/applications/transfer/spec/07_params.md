<!--
order: 7
-->

# Parameters

The ibc-transfer module contains the following parameters:

| Key              | Type | Default Value |
|------------------|------|---------------|
| `SendEnabled`    | bool | `true`        |
| `ReceiveEnabled` | bool | `true`        |

## SendEnabled

The transfers enabled parameter controls send cross-chain transfer capabilities for all fungible
tokens.

To prevent a single token from being transferred from the chain, set the `SendEnabled` parameter to `true` and
then set the bank module's [`SendEnabled` parameter](./../../../../bank/spec/05_params.md#sendenabled) for
the denomination to `false`.

## ReceiveEnabled

The transfers enabled parameter controls receive cross-chain transfer capabilities for all fungible
tokens.

To prevent a single token from being transferred to the chain, set the `ReceiveEnabled` parameter to `true` and
then set the bank module's [`SendEnabled` parameter](./../../../../bank/spec/05_params.md#sendenabled) for
the denomination to `false`.
