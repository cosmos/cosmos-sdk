<!--
order: 2
-->

# Messages

In this section we describe the processing of the crisis messages and the
corresponding updates to the state.

## MsgVerifyInvariant

Blockchain invariants can be checked using the `MsgVerifyInvariant` message.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc7/proto/cosmos/crisis/v1beta1/tx.proto#L14-L22

This message is expected to fail if:

- the sender does not have enough coins for the constant fee
- the invariant route is not registered

This message checks the invariant provided, and if the invariant is broken it
panics, halting the blockchain. If the invariant is broken, the constant fee is
never deducted as the transaction is never committed to a block (equivalent to
being refunded). However, if the invariant is not broken, the constant fee will
not be refunded.
