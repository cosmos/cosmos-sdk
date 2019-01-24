# Tags

The staking module emits the following events/tags:

## EndBlocker

| Key                   | Value                                     |
|-----------------------|-------------------------------------------|
| action                | complete-unbonding\|complete-redelegation |
| delegator             | {delegatorAccountAddress}                 |
| source-validator      | {srcOperatorAddress}                      |
| destination-validator | {dstOperatorAddress}                      |

## Handlers

### MsgCreateValidator

| Key                   | Value                |
|-----------------------|----------------------|
| destination-validator | {dstOperatorAddress} |
| moniker               | {validatorMoniker}   |
| identity              | {validatorIdentity}  |

### MsgEditValidator

| Key                   | Value                |
|-----------------------|----------------------|
| destination-validator | {dstOperatorAddress} |
| moniker               | {validatorMoniker}   |
| identity              | {validatorIdentity}  |

### MsgDelegate

| Key                   | Value                                     |
|-----------------------|-------------------------------------------|
| delegator             | {delegatorAccountAddress}                 |
| destination-validator | {dstOperatorAddress}                      |

### MsgBeginRedelegate

| Key                   | Value                                     |
|-----------------------|-------------------------------------------|
| delegator             | {delegatorAccountAddress}                 |
| source-validator      | {srcOperatorAddress}                      |
| destination-validator | {dstOperatorAddress}                      |
| end-time [0]          | {delegationFinishTime}                    |

* [0] Time is formatted in the RFC3339 standard

### MsgUndelegate

| Key              | Value                     |
|------------------|---------------------------|
| delegator        | {delegatorAccountAddress} |
| source-validator | {srcOperatorAddress}      |
| end-time [0]     | {delegationFinishTime}    |

* [0] Time is formatted in the RFC3339 standard
