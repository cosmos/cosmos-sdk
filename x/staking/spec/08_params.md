<!--
order: 8
-->

# Parameters

The staking module contains the following parameters:

| Key               | Type             | Example           |
|-------------------|------------------|-------------------|
| UnbondingTime     | string (time ns) | "259200000000000" |
| MaxValidators     | uint16           | 100               |
| KeyMaxEntries     | uint16           | 7                 |
| HistoricalEntries | uint16           | 3                 |
| BondDenom         | string           | "stake"           |
| EpochInterval     | int64            | 10                |

> Note: If the `EpochInterval` is set to 1, the messages will be executed at the end of each block.
