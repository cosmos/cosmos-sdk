# Tags

The slashing module emits the following events/tags:

## BeginBlocker

| Type  | Attribute Key | Attribute Value             |
|-------|---------------|-----------------------------|
| slash | address       | {validatorConsensusAddress} |
| slash | power         | {validatorPower}            |
| slash | reason        | {slashReason}               |
| slash | jailed [0]    | {validatorConsensusAddress} |

- [0] Only included if the validator is jailed. 

| Type     | Attribute Key | Attribute Value             |
|----------|---------------|-----------------------------|
| liveness | address       | {validatorConsensusAddress} |
| liveness | missed_blocks | {missedBlocksCounter}       |
| liveness | height        | {blockHeight}               |

## Handlers

### MsgUnjail

| Type    | Attribute Key | Attribute Value |
|---------|---------------|-----------------|
| message | module        | slashing        |
| message | action        | unjail          |
| message | sender        | {senderAddress} |
