# Messages

In this section we describe the processing of the staking messages and the
corresponding updates to the state. All created/modified state objects
specified by each message are defined within [state.md](state.md). 

## MsgCreateValidator

A validator is created using the `MsgCreateValidator` message. 

```golang
type MsgCreateValidator struct {
    Description    Description
    Commission     Commission

    DelegatorAddr  sdk.AccAddress
    ValidatorAddr  sdk.ValAddress
    PubKey         crypto.PubKey
    Delegation     sdk.Coin
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


## MsgEditValidator

The `Description`, `CommissionRate` of a validator can be updated using the
`MsgEditCandidacy`.  

```golang
type MsgEditCandidacy struct {
    Description     Description
    ValidatorAddr   sdk.ValAddress
    CommissionRate  sdk.Dec
}
```

This message is expected to fail if: 

 - the initial `CommissionRate` is either negative or > `MaxRate`
 - the `CommissionRate` has already been updated within the previous 24 hours
 - the `CommissionRate` is > `MaxChangeRate`
 - the description fields are too large

This message stores the updated `Validator` object. 

## MsgDelegate

Within this message the delegator provides coins, and in return receives
some amount of their validator's (newly created) delegator-shares that are
assigned to `Delegation.Shares`. 

```golang
type MsgDelegate struct {
	DelegatorAddr sdk.AccAddress
	ValidatorAddr sdk.ValAddress
	Delegation    sdk.Coin
}
```

This message is expected to fail if: 

 - the validator is does not exist
 - the validator is jailed 

If an existing `Delegation` object for provided addresses does not already
exist than it is created as part of this message otherwise the existing
`Delegation` is updated to include the newly received shares. 

## MsgBeginUnbonding

The begin unbonding message allows delegators to undelegate their tokens from
validator. 

```golang
type MsgBeginUnbonding struct {
	DelegatorAddr sdk.AccAddress 
	ValidatorAddr sdk.ValAddress
	SharesAmount  sdk.Dec 
}
```

This message is expected to fail if: 

 - the delegation doesn't exist
 - the validator doesn't exist
 - the delegation has less shares than `SharesAmount`
 - existing `UnbondingDelegation` has maximum entries as defined by
   params.MaxEntries

When this message is processed the following actions occur:
 - validator's `DelegatorShares` and the delegation's `Shares` are both reduced
   by the message `SharesAmount`
 - calculate the token worth of the shares remove that amount tokens held
   within the validator 
 - with those removed tokens, if the validator is:
   - bonded - add them to an entry in `UnbondingDelegation` (create
     `UnbondingDelegation` if it doesn't exist) with a completion time a full
     unbonding period from the current time. Update pool shares to reduce
     BondedTokens and increase NotBondedTokens by token worth of the shares. 
   - unbonding - add them to an entry in `UnbondingDelegation` (create
     `UnbondingDelegation` if it doesn't exist) with the same completion time
      as the validator (`UnbondingMinTime`).
   - unbonded - then send the coins the message `DelegatorAddr`
 - if there are no more `Shares` in the delegation, then the delegation object
   is removed from the store
   - under this situation if the delegation is the validator's self-delegation 
     then also jail the validator. 

## MsgBeginRedelegate

The redelegation command allows delegators to instantly switch validators. Once
the unbonding period has passed, the redelegation is automatically completed in
the EndBlocker.

```golang
type MsgBeginRedelegate struct {
	DelegatorAddr    sdk.AccAddress 
	ValidatorSrcAddr sdk.ValAddress 
	ValidatorDstAddr sdk.ValAddress
	SharesAmount     sdk.Dec
}
```

This message is expected to fail if: 

 - the delegation doesn't exist
 - the source or destination validators don't exist
 - the delegation has less shares than `SharesAmount`
 - the source validator has a receiving redelegation which
   is not matured (aka. the redelegation may be transitive) 
 - existing `Redelegation` has maximum entries as defined by
   params.MaxEntries

When this message is processed the following actions occur:
 - the source validator's `DelegatorShares` and the delegations `Shares` are
   both reduced by the message `SharesAmount`
 - calculate the token worth of the shares remove that amount tokens held
   within the source validator. 
 - if the source validator is:
   - bonded - add an entry to the `Redelegation` (create
     `Redelegation` if it doesn't exist) with a completion time a full
     unbonding period from the current time. Update pool shares to reduce
     BondedTokens and increase NotBondedTokens by token worth of the shares
     (this may be effectively reversed in the next step however). 
   - unbonding - add an entry to the `Redelegation` (create `Redelegation` if
     it doesn't exist) with the same completion time as the validator
     (`UnbondingMinTime`).
   - unbonded - no action required in this step
 - Delegate the token worth to the destination validator, possibly moving 
   tokens back to the bonded state. 
 - if there are no more `Shares` in the source delegation, then the source
   delegation object is removed from the store
   - under this situation if the delegation is the validator's self-delegation
     then also jail the validator. 
