# Tags

The distribution module emits the following events/tags:

## Handlers

### MsgSetWithdrawAddress

| Key        | Value                     |
|------------|---------------------------|
| `action`   | `set_withdraw_address`    |
| `category` | `distribution`            |
| `sender`   | {delegatorAccountAddress} |

### MsgWithdrawDelegatorReward

| Key                | Value                       |
|--------------------|-----------------------------|
| `action`           | `withdraw_delegator_reward` |
| `category`         | `distribution`              |
| `rewards`          | {rewards}                   |
| `sender`           | {delegatorAccountAddress}   |
| `source-validator` | {srcOperatorAddress}        |

### MsgWithdrawValidatorCommission

| Key          | Value                           |
|--------------|---------------------------------|
| `action`     | `withdraw_validator_commission` |
| `category`   | `distribution`                  |
| `commission` | {commission}                    |
| `sender`     | {srcOperatorAddress}            |
