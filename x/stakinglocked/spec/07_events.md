<!--
order: 7
-->

# Events

The staking module emits the following events:

## EndBlocker

| Type                  | Attribute Key         | Attribute Value           |
| --------------------- | --------------------- | ------------------------- |
| complete_unbonding    | amount                | {totalUnbondingAmount}    |
| complete_unbonding    | validator             | {validatorAddress}        |
| complete_unbonding    | delegator             | {delegatorAddress}        |
| complete_redelegation | amount                | {totalRedelegationAmount} |
| complete_redelegation | source_validator      | {srcValidatorAddress}     |
| complete_redelegation | destination_validator | {dstValidatorAddress}     |
| complete_redelegation | delegator             | {delegatorAddress}        |

## Msg's

### MsgCreateValidator

| Type             | Attribute Key | Attribute Value    |
| ---------------- | ------------- | ------------------ |
| create_validator | validator     | {validatorAddress} |
| create_validator | amount        | {delegationAmount} |
| message          | module        | staking            |
| message          | action        | create_validator   |
| message          | sender        | {senderAddress}    |

### MsgEditValidator

| Type           | Attribute Key       | Attribute Value     |
| -------------- | ------------------- | ------------------- |
| edit_validator | commission_rate     | {commissionRate}    |
| edit_validator | min_self_delegation | {minSelfDelegation} |
| message        | module              | staking             |
| message        | action              | edit_validator      |
| message        | sender              | {senderAddress}     |

### MsgDelegate

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| delegate | validator     | {validatorAddress} |
| delegate | amount        | {delegationAmount} |
| message  | module        | staking            |
| message  | action        | delegate           |
| message  | sender        | {senderAddress}    |

### MsgUndelegate

| Type    | Attribute Key       | Attribute Value    |
| ------- | ------------------- | ------------------ |
| unbond  | validator           | {validatorAddress} |
| unbond  | amount              | {unbondAmount}     |
| unbond  | completion_time [0] | {completionTime}   |
| message | module              | staking            |
| message | action              | begin_unbonding    |
| message | sender              | {senderAddress}    |

* [0] Time is formatted in the RFC3339 standard

### MsgCancelUnbondingDelegation

| Type                          | Attribute Key       | Attribute Value                     |
| ----------------------------- | ------------------  | ------------------------------------|
| cancel_unbonding_delegation   | validator           | {validatorAddress}                  |
| cancel_unbonding_delegation   | delegator           | {delegatorAddress}                  |
| cancel_unbonding_delegation   | amount              | {cancelUnbondingDelegationAmount}   |
| cancel_unbonding_delegation   | creation_height     | {unbondingCreationHeight}           |
| message                       | module              | staking                             |
| message                       | action              | cancel_unbond                       |
| message                       | sender              | {senderAddress}                     |

### MsgBeginRedelegate

| Type       | Attribute Key         | Attribute Value       |
| ---------- | --------------------- | --------------------- |
| redelegate | source_validator      | {srcValidatorAddress} |
| redelegate | destination_validator | {dstValidatorAddress} |
| redelegate | amount                | {unbondAmount}        |
| redelegate | completion_time [0]   | {completionTime}      |
| message    | module                | staking               |
| message    | action                | begin_redelegate      |
| message    | sender                | {senderAddress}       |

* [0] Time is formatted in the RFC3339 standard
