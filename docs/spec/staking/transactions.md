# Transaction Overview

In this section we describe the processing of the transactions and the
corresponding updates to the state.

Notes:
 - `tx` denotes a reference to the transaction being processed
 - `sender` denotes the address of the sender of the transaction
 - `getXxx`, `setXxx`, and `removeXxx` functions are used to retrieve and
   modify objects from the store
 - `sdk.Dec` refers to a decimal type specified by the SDK.
 - `sdk.Int` refers to an integer type specified by the SDK.
 

## TxCreateValidator

 - triggers: `distribution.CreateValidatorDistribution`

```golang
type TxCreateValidator struct {
    Description    Description
    Commission     Commission

    DelegatorAddr  sdk.AccAddress
    ValidatorAddr  sdk.ValAddress
    PubKey         crypto.PubKey
    Delegation     sdk.Coin
}
```

A validator is created using the `TxCreateValidator` transaction. 
This transaction is expected to fail if: 
 - another validator with this operator address is already registered
 - another validator with this pubkey is already registered
 - the initial self-delegation tokens are of a denom not specified as the
   bonding denom 
 - the commission parameters are faulty, namely:
   - `MaxRate` is either > 1 or < 0 
   - The initial `Rate` is either negative or > `MaxRate`
   - The initial `MaxChangeRate` is either negative or > `MaxRate`
 
This transaction creates and stores the `Validator` object at appropriate
indexes.  Additionally a self-delegation is made with the inital tokens
delegation tokens `Delegation`.  the validator always starts as unbonded but
may be bonded in the first end-block. 


## TxEditValidator

If either the `Description`, `Commission`, or the `ValidatorAddr` need to be
updated, the `TxEditCandidacy` transaction should be sent from the operator
account:

```golang
type TxEditCandidacy struct {
    Description     Description
    ValidatorAddr   sdk.ValAddress
    CommissionRate  sdk.Dec
}
```

editCandidacy(tx TxEditCandidacy):
    validator, ok := getValidator(tx.ValidatorAddr)
    if !ok return err // validator must exist

    // Attempt to update the validator's description. The description provided
    // must be valid.
    description, err := updateDescription(validator, tx.Description)
    if err != nil return err

    // a validator is not required to update it's commission rate
    if tx.CommissionRate != nil {
        // Attempt to update a validator's commission rate. The rate provided
        // must be valid. It's rate can only be updated once a day.
        err := updateValidatorCommission(validator, tx.CommissionRate)
        if err != nil return err
    }

    // set the validator and public key
    setValidator(validator)

    tags := createTags(tx)
    return tags

### TxDelegate

 - triggers: `distribution.CreateOrModDelegationDistribution`

Within this transaction the delegator provides coins, and in return receives
some amount of their validator's delegator-shares that are assigned to
`Delegation.Shares`.

```golang
type TxDelegate struct {
	DelegatorAddr sdk.Address
	ValidatorAddr sdk.Address
	Amount        sdk.Coin
}

delegate(tx TxDelegate):
    pool = getPool()
    if validator.Status == Jailed return

    delegation = getDelegatorBond(DelegatorAddr, ValidatorAddr)
    if delegation == nil then delegation = NewDelegation(DelegatorAddr, ValidatorAddr)

    validator, pool, issuedDelegatorShares = validator.addTokensFromDel(tx.Amount, pool)
    delegation.Shares += issuedDelegatorShares

    setDelegation(delegation)
    updateValidator(validator)
    setPool(pool)
    return
```

### TxStartUnbonding

Delegator unbonding is defined with the following transaction:

```golang
type TxStartUnbonding struct {
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

### TxRedelegation

The redelegation command allows delegators to instantly switch validators. Once
the unbonding period has passed, the redelegation is automatically completed in the EndBlocker.

```golang
type TxRedelegate struct {
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
