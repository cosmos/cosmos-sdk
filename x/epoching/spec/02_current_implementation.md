<!--
order: 2
-->

# Current Implementation

## Messages queue

Messages are queued to run at the end of epochs.
Queued messages have epoch number to be run and at the end of epochs, it run messages queued for the epoch and execute the message.

### Staking messages
- MsgCreateValidator
- MsgEditValidator
- MsgDelegate
- MsgBeginRedelegate
- MsgUndelegate

### Slashing messages
- MsgUnjail

### Evidence messages
- MsgSubmitEvidence

### Distribution messages
- MsgSetWithdrawAddress
- MsgWithdrawDelegatorReward
- MsgWithdrawValidatorCommission
- MsgFundCommunityPool

## Execution on epochs
- Try executing the message for the epoch
- If success, make changes as it is
- If failure, try making revert extra actions done on handlers (e.g. EpochTempPool deposit)
- If revert fail, panic
