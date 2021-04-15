<!--
order: 5
-->

# Hooks

## Create or modify delegation distribution

 - triggered-by: `staking.MsgDelegate`, `staking.MsgBeginRedelegate`, `staking.MsgUndelegate`

### Before

The delegation rewards are withdrawn to the delegator's withdraw address.
The rewards include the current period, but excludes the starting period.
Then increment the validator period, which is used for calculating rewards.
The latter is necessary because, due to the nature of the triggers, the validator's power and share distribution can have changed.
The reference count for the delegator's starting period is decremented.

### After

The starting height of the delegation is set to the previous period.
Because of the `Before`-hook, this is the last period for which the delegator was rewarded.

## Validator created

- triggered-by: `staking.MsgCreateValidator`

Initialized the validator variables: Historical rewards, current accumulated rewards, accumulated commission, and total outstanding rewards.
All are set to a `0` by default.
The period is set to `1`.

## Validator removed

- triggered-by: `staking.RemoveValidator`

Any outstanding commission is sent to the validator's self-delegation's withdrawal address.
Any remaining rewards still stored in the validator gets sent to the community fee pool.

Note: the validator only gets removed when it has no remaining delegations.
At that time, all outstanding delegator rewards will have been withdrawn.
Any remaining rewards are dust amounts.

## Validator is slashed

- triggered-by: `staking.Slash`
  
The current validator period reference count is incremented, since the slash event will need to refer to it.
The validator period is incremented.
The slash event is stored, to be referenced when calculating delegator rewards.
