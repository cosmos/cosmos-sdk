# Tags

The distribution module emits the following events/tags:

## Handlers

### MsgSetWithdrawAddress

| Key       | Value                     |
| --------- | ------------------------- |
| delegator | {delegatorAccountAddress} |

### MsgWithdrawDelegatorReward

| Key              | Value                     |
| ---------------- | ------------------------- |
| rewards          | {rewards}                 |
| delegator        | {delegatorAccountAddress} |
| source-validator | {srcOperatorAddress}      |

### MsgWithdrawValidatorRewardsAll

| Key              | Value                |
| ---------------- | -------------------- |
| commission       | {commission}         |
| source-validator | {srcOperatorAddress} |
