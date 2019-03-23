# Messages

In this section we describe the processing of the crisis messages and the
corresponding updates to the state. 

## MsgVerifyInvariant

Blockchain Invariance can be check using the `MsgVerifyInvariant` message. 

```golang
type MsgCreateValidator struct {
	Sender         sdk.AccAddress 
	InvariantRoute string
}
```

This message is expected to fail if: 

 - another validator with this operator address is already registered
 - another validator with this pubkey is already registered
 - the initial self-delegation tokens are of a denom not specified as the
   bonding denom 
 - the commission parameters are faulty, namely:
   - `MaxRate` is either > 1 or < 0 
   - the initial `Rate` is either negative or > `MaxRate`
   - the initial `MaxChangeRate` is either negative or > `MaxRate`
 - the description fields are too large
 
This message creates and stores the `Validator` object at appropriate indexes.
Additionally a self-delegation is made with the initial tokens delegation
tokens `Delegation`. The validator always starts as unbonded but may be bonded
in the first end-block. 


