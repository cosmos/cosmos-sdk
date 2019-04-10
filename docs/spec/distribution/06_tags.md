# Tags

The distribution module emits the following events/tags:

## Handlers

### MsgSetWithdrawAddress

| Key         | Value                     |
|-------------|---------------------------|
| `action`    | `set_withdraw_address`    |
| `category`  | `distribution`            |
| `delegator` | {delegatorAccountAddress} |

### MsgWithdrawDelegatorReward

| Key                | Value                       |
|--------------------|-----------------------------|
| `action`           | `withdraw_delegator_reward` |
| `category`         | `distribution`              |
| `delegator`        | {delegatorAccountAddress}   |
| `source-validator` | {srcOperatorAddress}        |

### MsgWithdrawValidatorCommission

| Key                | Value                           |
|--------------------|---------------------------------|
| `action`           | `withdraw_validator_commission` |
| `category`         | `distribution`                  |
| `source-validator` | {srcOperatorAddress}            |
