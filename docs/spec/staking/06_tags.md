# Tags

The staking module emits the following events/tags:

## EndBlocker

| Key                     | Value                                         |
|-------------------------|-----------------------------------------------|
| `action`                | `complete-unbonding`\|`complete-redelegation` |
| `category`              | `staking`                                     |
| `delegator`             | {delegatorAccountAddress}                     |
| `source-validator`      | {srcOperatorAddress}                          |
| `destination-validator` | {dstOperatorAddress}                          |

## Handlers

### MsgCreateValidator

| Key        | Value                |
|------------|----------------------|
| `action`   | `create_validator`   |
| `category` | `staking`            |
| `sender`   | {dstOperatorAddress} |

### MsgEditValidator

| Key        | Value                |
|------------|----------------------|
| `action`   | `edit_validator`     |
| `category` | `staking`            |
| `sender`   | {dstOperatorAddress} |

### MsgDelegate

| Key                     | Value                     |
|-------------------------|---------------------------|
| `action`                | `delegate`                |
| `category`              | `staking`                 |
| `sender`                | {delegatorAccountAddress} |
| `destination-validator` | {dstOperatorAddress}      |

### MsgBeginRedelegate

| Key                     | Value                     |
|-------------------------|---------------------------|
| `action`                | `begin_redelegate`        |
| `category`              | `staking`                 |
| `sender`                | {delegatorAccountAddress} |
| `source-validator`      | {srcOperatorAddress}      |
| `destination-validator` | {dstOperatorAddress}      |
| `end-time` [0]          | {delegationFinishTime}    |

* [0] Time is formatted in the RFC3339 standard

### MsgUndelegate

| Key                | Value                     |
|--------------------|---------------------------|
| `action`           | `begin_unbonding`         |
| `category`         | `staking`                 |
| `sender`           | {delegatorAccountAddress} |
| `source-validator` | {srcOperatorAddress}      |
| `end-time` [0]     | {delegationFinishTime}    |

* [0] Time is formatted in the RFC3339 standard
