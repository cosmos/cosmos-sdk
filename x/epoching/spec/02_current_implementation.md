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

All `slashing` module messages are queued. (Only one message btw :) )
### Evidence messages
- **MsgSubmitEvidence**: No changes

No messages are queued on `evidence` module for now.

### Distribution messages
- **MsgSetWithdrawAddress**: No changes
- **MsgWithdrawDelegatorReward**: No changes
- **MsgWithdrawValidatorCommission**: No changes
- **MsgFundCommunityPool**: No changes

No messages are queued on `distribution` module for now.

## Slash and Jail

Slash and Jail is automatically done on BeginBlocker / Endblocker.
Currently validator set update is only done on staking module's endblocker and other modules'(which affect Slash / Jail) Endblockers are being executed before staking module.

For now, Slash and Jail take effect instantly at the end of block.

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

## Discussions / Review / Research for epoching

```go
  // Current logic flow summary:
  // Current queues: "unbonding validator", "unbonding delegator", “redelegation"
  // Current Queue management
  // EndBlocker -> BlockValidatorUpdates -> DequeueAllMatureUBDQueue
  // EndBlocker -> BlockValidatorUpdates -> DequeueAllMatureRedelegationQueue
  // EndBlocker -> BlockValidatorUpdates -> UnbondAllMatureValidators 
  // EndBlocker -> BlockValidatorUpdates -> CompleteUnbonding -> RemoveUnbondingDelegation
  // EndBlocker -> BlockValidatorUpdates -> CompleteRedelegation -> RemoveRedelegation
  // BeginBlocker -> HandleValidatorSignature -> Jail
  // BeginBlocker -> HandleValidatorSignature -> Slash

  // Target logic flow summary:
  // Epoch -> BlockValidatorUpdates -> DequeueAllMatureUBDQueue
  // Epoch -> BlockValidatorUpdates -> DequeueAllMatureRedelegationQueue
  // Epoch -> BlockValidatorUpdates -> UnbondAllMatureValidators 
  // Epoch -> BlockValidatorUpdates -> CompleteUnbonding -> RemoveUnbondingDelegation
  // Epoch -> BlockValidatorUpdates -> CompleteRedelegation -> RemoveRedelegation
  // BeginBlocker -> HandleValidatorSignature -> Jail
  // BeginBlocker -> HandleValidatorSignature -> Slash

  // Current Msgs management
  // MsgUnjail -> instant validator update on condition met
  // MsgCreateValidator -> instant validator creation
  // MsgEditValidator -> instant validator update
  // MsgBeginRedelegate -> instant validator update and queue redelegation
  // MsgUndelegate -> instant pool balance update and queue undelegate
  // MsgDelegate -> instant delegation update
  // MsgWithdrawValidatorCommission -> instant
  // MsgWithdrawDelegatorReward -> instant

  // Target Msgs management
  // MsgUnjail -> queue validator update on condition met (BufferedMsgUnjailQueue)
  // MsgCreateValidator -> queue validator creation on condition met (BufferedMsgCreateValidatorQueue)
  // MsgEditValidator -> queue validator edit on condition met (BufferedMsgEditValidatorQueue)
  // MsgDelegate -> queue delegate (BufferedMsgDelegateQueue)
  // MsgBeginRedelegate -> queue redelegation (BufferedMsgRedelegationQueue) => move tokens between validators on epoch => After 3 weeks time, it should automatically remove redelegation entry for completion even though it's nott the end of epoch
  // MsgUndelegate -> queue undelegation (BufferedMsgUndelegateQueue) => move tokens to NotBondedPool and start unbonding period on epoch => After 3 weeks time, it should automatically unbond even though it’s not the end of epoch
  // MsgWithdrawValidatorCommission -> instant
  // MsgWithdrawDelegatorReward -> instant

  // The flow for an unbonding process would be:

  // 1. Submit MsgUnbond which adds it to DelegationChangeQueue
  // 2. Wait for end of Epoch
  // 3. Execute "BeginUnbonding", this adds it to UnbondingQueue
  // 4. Wait till end of Unbonding Period (3 weeks)
  // 5. Remove from UnbondingQueue

  // When a validator begins the unbonding process, it turns the validator into unbonding state instantly.
  // This is different than a specific delegator beginning to unbond. A validator beginning to unbond means that it's not in the set any more.
  // A delegator unbonding from a validator removes their delegation from the validator.

  // Cases that trigger unbonding process
  // - Validator undelegate can unbond more tokens than his minimum_self_delegation and it will automatically turn the validator into unbonding
  //   In this case, unbonding should start instantly.
  // - Validator miss blocks and get slashed
  // - Validator get slashed for double sign
  
  // The order of running buffered msgs on epoching could affect something?
  // e.g. MsgUnjail could happen later time than MsgDelegate and next Jail/Slash event.
  // e.g. MsgUnjail and MsgUndelegate could happen in different order. MsgUndelegate after MsgUnjail.
  // I think it won't affect anything, btw, here's current ordering of implementation in simapp.
  // 	app.mm.SetOrderBeginBlockers(
	// 	upgradetypes.ModuleName, minttypes.ModuleName, distrtypes.ModuleName, slashingtypes.ModuleName,
	// 	evidencetypes.ModuleName, stakingtypes.ModuleName, ibchost.ModuleName,
	// )
	// app.mm.SetOrderEndBlockers(crisistypes.ModuleName, govtypes.ModuleName, stakingtypes.ModuleName)

  // Changes to make for logic flow
  // BlockValidatorUpdates should be modified to ValidatorUpdates and should be called on Epoch
  // BeginBlocker -> HandleValidatorSignature run as it is and run ValidatorUpdates when it's updated on EndBlocker
  // to trigger tendermint validator set update instantly when Jail/Slash case.

  // Changes to make for Msgs flow
  // BufferMsgUnjail should be added to execute on next epoch: Add when condition met
  // BufferMsgCreateValidator should be added: send from user account to NotBondedPool and queue action
  // BufferMsgEditValidator should be added: queue action
  // BufferMsgDelegate should be added: queue action
  // BufferMsgBeginRedelegate should be added: queue action
  // BufferMsgUnDelegate should be added: queue action
```