<!--
order: 6
-->

# Parameters

The ibc-transfer module contains the following parameters:

| Key                | Type | Default Value |
|--------------------|------|---------------|
| `TransfersEnabled` | bool | `true`        |

## TransfersEnabled

The transfers enabled parameter controls send and receive cross-chain transfer capabilities for all
fungible tokens.

To prevent a single token from being transferred, set the `TransfersEnabled` parameter to `true` and
then set the bank module's [`SendEnabled` parameter](./../../bank/spec/05_params.md#sendenabled) for
the denomination to `false`.
