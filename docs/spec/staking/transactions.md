# Transaction Overview

In this section we describe the processing of the transactions and the
corresponding updates to the state.

## MsgCreateValidator

A validator is created using the `MsgCreateValidator` transaction. 

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

This transaction is expected to fail if: 

 - another validator with this operator address is already registered
 - another validator with this pubkey is already registered
 - the initial self-delegation tokens are of a denom not specified as the
   bonding denom 
 - the commission parameters are faulty, namely:
   - `MaxRate` is either > 1 or < 0 
   - the initial `Rate` is either negative or > `MaxRate`
   - the initial `MaxChangeRate` is either negative or > `MaxRate`
 - the description fields are too large
 
This transaction creates and stores the `Validator` object at appropriate
indexes.  Additionally a self-delegation is made with the inital tokens
delegation tokens `Delegation`.  the validator always starts as unbonded but
may be bonded in the first end-block. 


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

This transaction is expected to fail if: 

 - the initial `CommissionRate` is either negative or > `MaxRate`
 - the `CommissionRate` has already been updated within the previous 24 hours
 - the `CommissionRate` is > `MaxChangeRate`
 - the description fields are too large

This transaction stores the updated `Validator` object. 

### MsgDelegate

Within this transaction the delegator provides coins, and in return receives
some amount of their validator's (newly created) delegator-shares that are
assigned to `Delegation.Shares`. 

```golang
type MsgDelegate struct {
	DelegatorAddr sdk.Address
	ValidatorAddr sdk.Address
	Amount        sdk.Coin
}
```

This transaction is expected to fail if: 

 - the validator is does not exist
 - the validator is jailed 

If an existing `Delegation` object for provided addresses does not already
exist than it is created as part of this transaction otherwise the existing
`Delegation` is updated to include the newly received shares. 

### MsgBeginUnbonding

Delegator unbonding is defined with the following transaction:

```golang
type MsgBeginUnbonding struct {
	DelegatorAddr sdk.Address
	ValidatorAddr sdk.Address
	Shares        string
}

startUnbonding(tx TxStartUnbonding):
    delegation, found = getDelegatorBond(store, sender, tx.PubKey)
    if !found == nil return

		if bond.Shares < tx.Shares
			return ErrNotEnoughBondShares

	validator, found = GetValidator(tx.ValidatorAddr)
	if !found {
		return err

	bond.Shares -= tx.Shares

	revokeCandidacy = false
	if bond.Shares.IsZero() {

		if bond.DelegatorAddr == validator.Operator && validator.Jailed == false
			revokeCandidacy = true

		removeDelegation( bond)
	else
		bond.Height = currentBlockHeight
		setDelegation(bond)

	pool = GetPool()
	validator, pool, returnAmount = validator.removeDelShares(pool, tx.Shares)
	setPool( pool)

    unbondingDelegation = NewUnbondingDelegation(sender, returnAmount, currentHeight/Time, startSlashRatio)
    setUnbondingDelegation(unbondingDelegation)

	if revokeCandidacy
		validator.Jailed = true

	validator = updateValidator(validator)

	if validator.Status == Unbonded && validator.DelegatorShares == 0 {
		removeValidator(validator.Operator)

    return
```

### MsgBeginRedelegate

The redelegation command allows delegators to instantly switch validators. Once
the unbonding period has passed, the redelegation is automatically completed in
the EndBlocker.

```golang
type MsgBeginRedelegate struct {
    DelegatorAddr Address
    ValidatorFrom Validator
    ValidatorTo   Validator
    Shares        sdk.Dec 
    CompletedTime int64 
}

redelegate(tx TxRedelegate):

    pool = getPool()
    delegation = getDelegatorBond(tx.DelegatorAddr, tx.ValidatorFrom.Operator)
    if delegation == nil
        return

    if delegation.Shares < tx.Shares
        return
    delegation.shares -= Tx.Shares
    validator, pool, createdCoins = validator.RemoveShares(pool, tx.Shares)
    setPool(pool)

    redelegation = newRedelegation(tx.DelegatorAddr, tx.validatorFrom,
        tx.validatorTo, tx.Shares, createdCoins, tx.CompletedTime)
    setRedelegation(redelegation)
    return
```
