# Events

The POA module emits the following events:

## EndBlocker

| Type               | Attribute Key | Attribute Value    |
| ------------------ | ------------- | ------------------ |
| complete_unbonding | validator     | {validatorAddress} |

## Handlers

### MsgCreateValidator

| Type             | Attribute Key | Attribute Value    |
| ---------------- | ------------- | ------------------ |
| create_validator | validator     | {validatorAddress} |
| create_validator | weight        | 10                 |
| message          | module        | poa                |
| message          | action        | create_validator   |
| message          | sender        | {senderAddress}    |

### MsgProposeCreateValidator

| Type             | Attribute Key | Attribute Value    |
| ---------------- | ------------- | ------------------ |
| create_validator | validator     | {validatorAddress} |
| create_validator | weight        | 10                 |
| message          | module        | poa                |
| message          | action        | create_validator   |
| message          | sender        | {senderAddress}    |

### MsgProposeNewWeight

| Type             | Attribute Key | Attribute Value    |
| ---------------- | ------------- | ------------------ |
| create_validator | validator     | {validatorAddress} |
| create_validator | newWeight     | {newWeight}        |
| message          | module        | poa                |
| message          | action        | new_weight         |
| message          | sender        | {senderAddress}    |

### MsgEditValidator

| Type    | Attribute Key | Attribute Value |
| ------- | ------------- | --------------- |
| message | module        | poa             |
| message | action        | edit_validator  |
| message | sender        | {senderAddress} |
