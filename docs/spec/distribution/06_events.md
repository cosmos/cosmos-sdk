# Events

The distribution module emits the following events:

## BeginBlocker

| Type            | Attribute Key | Attribute Value    |
|-----------------|---------------|--------------------|
| proposer_reward | validator     | {validatorAddress} |
| proposer_reward | reward        | {proposerReward}   |
| commission      | amount        | {commissionAmount} |
| commission      | validator     | {validatorAddress} |
| rewards         | amount        | {rewardAmount}     |
| rewards         | validator     | {validatorAddress} |

## Handlers

### MsgSetWithdrawAddress

| Type                 | Attribute Key    | Attribute Value      |
|----------------------|------------------|----------------------|
| set_withdraw_address | sender           | {senderAddress}      |
| set_withdraw_address | withdraw_address | {withdrawAddress}    |
| message              | module           | distribution         |
| message              | action           | set_withdraw_address |

### MsgWithdrawDelegatorReward

| Type    | Attribute Key | Attribute Value           |
|---------|---------------|---------------------------|
| rewards | amount        | {rewardAmount}            |
| rewards | sender        | {senderAddress}           |
| rewards | validator     | {validatorAddress}        |
| message | module        | distribution              |
| message | action        | withdraw_delegator_reward |

### MsgWithdrawValidatorCommission

| Type       | Attribute Key | Attribute Value               |
|------------|---------------|-------------------------------|
| commission | amount        | {commissionAmount}            |
| commission | sender        | {senderAddress}               |
| message    | module        | distribution                  |
| message    | action        | withdraw_validator_commission |
