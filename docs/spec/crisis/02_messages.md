# Messages

In this section we describe the processing of the crisis messages and the
corresponding updates to the state. 

## MsgVerifyInvariant

Blockchain invariants can be check using the `MsgVerifyInvariant` message. 

```golang
type MsgVerifyInvariant struct {
	Sender         sdk.AccAddress 
	InvariantRoute string
}
```

This message is expected to fail if: 
 - the sender does not have enough coins for the constant fee
 - the invariant route is not registered 

This message checks the invariant provided, and if the invariant is broken it
panics, halting the blockchain. If the invariant is broken, the constant fee is
refunded to the message sender from the community pool, however if the
invariant is not broken, the constant fee will not be refunded.

This message should never fail if the invariant is broken, even if there are
insufficient funds to refund the sender of the message. 
