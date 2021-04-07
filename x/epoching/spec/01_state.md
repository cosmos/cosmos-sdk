<!--
order: 1
-->

# State

## Messages queue

Messages are queued to run at the end of epochs. Queued messages have epoch number to be run and at the end of epochs, it run messages queued for the epoch and execute the message.

### Message queues

Each module has 1 message queue that is specific to a module.

## Actions

A module will add a message that complies with the `sdk.Msg` interface. These message will be executed at a later date.

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

For now, it's implemented to export all buffered messages without epoch number. And when import, Buffered messages are stored on current epoch to run at the end of current epoch.

## Genesis Transactions

We execute epoch after execution of genesis transactions to see the changes instantly before node start.

## Execution on epochs

- Try executing the message for the epoch
- If success, make changes as it is
- If failure, try making revert extra actions done on handlers (e.g. EpochDelegationPool deposit)
- If revert fail, panic
