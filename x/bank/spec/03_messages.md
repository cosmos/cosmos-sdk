<!--
order: 3
-->

# Messages

## MsgSend

Send coins from one address to another.
+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/bank/v1beta1/tx.proto#L19-L28

The message will fail under the following conditions:

- The coins do not have sending enabled
- The `to` address is restricted
