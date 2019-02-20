# Tags

The distribution module emits the following events/tags:

## Handlers

### MsgSetWithdrawAddress

| Key       | Value                     |
|-----------|---------------------------|
| delegator | {delegatorAccountAddress} |

### MsgWithdrawDelegatorRewardsAll

| Key       | Value                     |
|-----------|---------------------------|
| delegator | {delegatorAccountAddress} |

### MsgWithdrawDelegatorReward

| Key              | Value                     |
|------------------|---------------------------|
| delegator        | {delegatorAccountAddress} |
| source-validator | {srcOperatorAddress}      |

### MsgWithdrawValidatorRewardsAll

| Key              | Value                |
|------------------|----------------------|
| source-validator | {srcOperatorAddress} |
