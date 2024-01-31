<!--
order: 3
-->

# Messages

## MsgSend

Send coins from one address to another.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/bank/v1beta1/tx.proto#L21-L32

The message will fail under the following conditions:

* The coins do not have sending enabled
* The `to` address is restricted

## MsgMultiSend

Send coins from and to a series of different address. If any of the receiving addresses do not correspond to an existing account, a new account is created.
+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/proto/cosmos/bank/v1beta1/tx.proto#L37-L45

The message will fail under the following conditions:

* Any of the coins do not have sending enabled
* Any of the `to` addresses are restricted
* Any of the coins are locked
* The inputs and outputs do not correctly correspond to one another
