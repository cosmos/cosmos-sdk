<!--
order: 2
-->

# Current Implementation

## Messages queue

Messages are queued to run at the end of epochs.
Queued messages have epoch number to be run and at the end of epochs, it run messages queued for the epoch and execute the message.

### Staking messages
- **MsgCreateValidator**: Move user's funds to `EpochDelegationPool` inside handler and move funds in `EpochDelegationPool` to `UnbondedPool` on Epoch while doing self delegation. If Epoch execution fail, return back funds from `EpochDelegationPool` to user's account.
- **MsgEditValidator**: Validate message and if valid queue the message for execution at the end of the Epoch.
- **MsgDelegate**: Move user's funds to `EpochDelegationPool` inside handler and move funds in `EpochDelegationPool` to `UnbondedPool` on Epoch while doing delegation to a selected validator. If Epoch execution fail, return back funds from `EpochDelegationPool` to user's account.
- **MsgBeginRedelegate**: Validate message and if valid queue the message for execution at the end of the Epoch.
- **MsgUndelegate**: Validate message and if valid queue the message for execution at the end of the Epoch.

All `staking` module messages are queued.
### Message queues

Each module has 1 message queue. Currently, there are two queues, one for `staking` and the other for `slashing`.
Each `module` message queue, saves the queued messages for the module.

### Slashing messages
- **MsgUnjail**: Validate message and if valid queue the message for execution at the end of the Epoch.

All `slashing` module messages are queued.

Note: `SlashEvent` execution is also put on slashing message queue and executed at the end of the current epoch.
### Evidence messages
- **MsgSubmitEvidence**: No changes

No messages are queued on `evidence` module for now.

### Distribution messages
- **MsgSetWithdrawAddress**: No changes
- **MsgWithdrawDelegatorReward**: No changes
- **MsgWithdrawValidatorCommission**: No changes
- **MsgFundCommunityPool**: No changes

No messages are queued on `distribution` module for now.

## Slash and Jail on slashing/evidence module

Jail action is automatically executed at the end of current block.
Slash action is queued and executed at the end of current epoch.

Currently validator set update is only done on staking module's endblocker and slashing/evidence module Endblockers are being executed before staking module.

No changes were made on `evidence` module since it's related to `Jail` which requires instant action.

## Execution on epochs
- Try executing the message for the epoch
- If success, make changes as it is
- If failure, try making revert extra actions done on handlers (e.g. EpochDelegationPool deposit)
- If revert fail, panic

## Endblocker ValidatorSetUpdates

Validator set update is done on every block to care about `Jailed` validators.
`Jailed` validator should take effect instantly but rest should take effect at the end of the epoch.

## Buffered Messages Export / Import

For now, it's implemented to export all buffered messages without epoch number. And when import, Buffered messages are stored on current epoch to run at the end of current epoch.

## Genesis Transactions

We execute epoch after execution of genesis transactions to see the changes instantly before node start.

## Flow for unbonding process

1. Submit MsgUnbond which adds it to EpochingMsgQueue (staking module)
2. Wait for end of Epoch
3. Execute "BeginUnbonding", this adds it to UnbondingQueue
4. Wait till end of Unbonding Period (3 weeks)
5. Remove from UnbondingQueue

## Order of epoch execution

Staking module's endblocker come at the end as `Validator Set` update is done at the endblocker of staking module.
