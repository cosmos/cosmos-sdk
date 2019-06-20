# Events

The staking module emits the following events:

## EndBlocker

| Type                  | Attribute Key         | Attribute Value       |
|-----------------------|-----------------------|-----------------------|
| complete_unbonding    | validator             | {validatorAddress}    |
| complete_unbonding    | delegator             | {delegatorAddress}    |
| complete_redelegation | source_validator      | {srcValidatorAddress} |
| complete_redelegation | destination_validator | {dstValidatorAddress} |
| complete_redelegation | delegator             | {delegatorAddress}    |                        |

## Handlers

### MsgCreateValidator

| Type             | Attribute Key | Attribute Value    |
|------------------|---------------|--------------------|
| create_validator | sender        | {senderAddress}    |
| create_validator | validator     | {validatorAddress} |
| create_validator | amount        | {delegationAmount} |
| message          | module        | staking            |
| message          | action        | create_validator   |

### MsgEditValidator

| Type           | Attribute Key | Attribute Value |
|----------------|---------------|-----------------|
| edit_validator | sender        | {senderAddress} |
| message        | module        | staking         |
| message        | action        | edit_validator  |

### MsgDelegate

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| delegate | sender        | {senderAddress}    |
| delegate | validator     | {validatorAddress} |
| delegate | amount        | {delegationAmount} |
| message  | module        | staking            |
| message  | action        | delegate           |

### MsgUndelegate

| Type    | Attribute Key   | Attribute Value    |
|---------|-----------------|--------------------|
| unbond  | sender          | {senderAddress}    |
| unbond  | validator       | {validatorAddress} |
| unbond  | amount          | {unbondAmount}     |
| unbond  | completion_time [0] | {completionTime}   |
| message | module          | staking            |
| message | action          | begin_unbonding    |

* [0] Time is formatted in the RFC3339 standard

### MsgBeginRedelegate

| Type       | Attribute Key         | Attribute Value       |
|------------|-----------------------|-----------------------|
| redelegate | sender                | {senderAddress}       |
| redelegate | source_validator      | {srcValidatorAddress} |
| redelegate | destination_validator | {dstValidatorAddress} |
| redelegate | amount                | {unbondAmount}        |
| redelegate | completion_time [0]   | {completionTime}      |
| message    | module                | staking               |
| message    | action                | begin_redelegate      |

* [0] Time is formatted in the RFC3339 standard
