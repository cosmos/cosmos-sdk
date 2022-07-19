<!--
order: 1
-->

# State

## Messages queue

Messages are queued to run at the end of each epoch. Queued messages have an epoch number and for each epoch number, the queues are iterated over and each message is executed.

### Message queues

Each module has one unique message queue that is specific to that module.

## Actions

A module will add a message that implements the `sdk.Msg` interface. These message will be executed at a later time (end of the next epoch).

```go
type Msg interface {
  proto.Message

  // Return the message type.
  // Must be alphanumeric or empty.
  Route() string

  // Returns a human-readable string for the message, intended for utilization
  // within tags
  Type() string

  // ValidateBasic does a simple validation check that
  // doesn't require access to any other information.
  ValidateBasic() error

  // Get the canonical byte representation of the Msg.
  GetSignBytes() []byte

  // Signers returns the addrs of signers that must sign.
  // CONTRACT: All signatures must be present to be valid.
  // CONTRACT: Returns addrs in some deterministic order.
  GetSigners() []AccAddress
 }
```

## Buffered Messages Export / Import

For now, the `x/epoching` module is implemented to export all buffered messages without epoch numbers. When state is imported, buffered messages are stored on current epoch to run at the end of current epoch.

## Genesis Transactions

We execute epoch after execution of genesis transactions to see the changes instantly before node start.

## Execution on epochs

* Try executing the message for the epoch
* If success, make changes as it is
* If failure, try making revert extra actions done on handlers (e.g. EpochDelegationPool deposit)
* If revert fail, panic
