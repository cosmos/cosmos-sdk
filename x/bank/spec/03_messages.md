<!--
order: 3
-->

# Messages

## MsgSend

Send coins from one address to another.
+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/bank/v1beta1/tx.proto#L19-L28

The message will fail under the following conditions:

* The coins do not have sending enabled
* The `to` address is restricted

## MsgMultiSend

Send coins from and to a series of different address. If any of the receiving addresses do not correspond to an existing account, a new account is created.
+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/bank/v1beta1/tx.proto#L33-L39

The message will fail under the following conditions:

* Any of the coins do not have sending enabled
* Any of the `to` addresses are restricted
* Any of the coins are locked
* The inputs and outputs do not correctly correspond to one another

## MsgSetSendEnabled

Used with the x/gov module to set create/edit SendEnabled entries.
+++ https://github.com/cosmos/cosmos-sdk/blob/dwedul/11859-send-disabled-change/proto/cosmos/bank/v1beta1/tx.proto#L53-L60

The message will fail under the following conditions:

* The authority is not a bech32 address.
* The authority is not x/gov module's address.
* There are multiple SendEnabled entries with the same Denom.
* One or more SendEnabled entries has an invalid Denom.
