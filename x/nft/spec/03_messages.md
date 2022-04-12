<!--
order: 3
-->

# Messages

In this section we describe the processing of messages for the nft module.

## MsgSend

You can use the `MsgSend` message to transfer the ownership of nft. This is a function provided by the `x/nft` module. Of course, you can use the `Transfer` method to implement your own transfer logic, but you need to pay extra attention to the transfer permissions.

The message handling should fail if:

* provided `ClassID` is not exist.
* provided `Id` is not exist.
* provided `Sender` is not the owner of nft.
