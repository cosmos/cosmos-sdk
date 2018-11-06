# Transaction Overview

In this section we describe the processing of the transactions and the
corresponding updates to the state. Transactions:

* TxCreateValidator
* TxEditValidator
* TxDelegation
* TxStartUnbonding
* TxRedelegate

Other important state changes:

* Update Validators

Other notes:

* `tx` denotes a reference to the transaction being processed
* `sender` denotes the address of the sender of the transaction
* `getXxx`, `setXxx`, and `removeXxx` functions are used to retrieve and
  modify objects from the store
* `sdk.Dec` refers to a decimal type specified by the SDK.

## TxCreateValidator

* triggers: `distribution.CreateValidatorDistribution`

A validator is created using the `TxCreateValidator` transaction.

```golang
type TxCreateValidator struct {
    Description    Description
    Commission     Commission

    DelegatorAddr  sdk.AccAddress
    ValidatorAddr  sdk.ValAddress
    PubKey         crypto.PubKey
    Delegation     sdk.Coin
}

createValidator(tx TxCreateValidator):
    ok := validatorExists(tx.ValidatorAddr)
    if ok return err // only one validator per address

    ok := validatorByPubKeyExists(tx.PubKey)
    if ok return err // only one validator per public key

    err := validateDenom(tx.Delegation.Denom)
    if err != nil return err // denomination must be valid

    validator := NewValidator(tx.ValidatorAddr, tx.PubKey, tx.Description)

    err := setInitialCommission(validator, tx.Commission, blockTime)
    if err != nil return err // must be able to set initial commission correctly

    // set the validator and public key
    setValidator(validator)
    setValidatorByPubKeyIndex(validator)

    // delegate coins from tx.DelegatorAddr to the validator
    err := delegate(tx.DelegatorAddr, tx.Delegation, validator)
    if err != nil return err // must be able to set delegation correctly

    tags := createTags(tx)
    return tags
```

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
```

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

### Update Validators

Within many transactions the validator set must be updated based on changes in
power to a single validator. This process also updates the Tendermint-Updates
store for use in end-block when validators are either added or kicked from the
Tendermint.

```golang
updateBondedValidators(newValidator Validator) (updatedVal Validator)

	kickCliffValidator = false
	oldCliffValidatorAddr = getCliffValidator(ctx)

	// add the actual validator power sorted store
	maxValidators = GetParams(ctx).MaxValidators
	iterator = ReverseSubspaceIterator(ValidatorsByPowerKey) // largest to smallest
	bondedValidatorsCount = 0
	var validator Validator
	for {
		if !iterator.Valid() || bondedValidatorsCount > int(maxValidators-1) {

			if bondedValidatorsCount == int(maxValidators) { // is cliff validator
				setCliffValidator(ctx, validator, GetPool(ctx))
			iterator.Close()
			break

		// either retrieve the original validator from the store,
		// or under the situation that this is the "new validator" just
		// use the validator provided because it has not yet been updated
		// in the main validator store

		operatorAddr = iterator.Value()
		if bytes.Equal(operatorAddr, newValidator.Operator) {
			validator = newValidator
        else
			validator = getValidator(operatorAddr)

		// if not previously a validator (and unjailed),
		// kick the cliff validator / bond this new validator
		if validator.Status() != Bonded && !validator.Jailed {
			kickCliffValidator = true

			validator = bondValidator(ctx, store, validator)
			if bytes.Equal(operatorAddr, newValidator.Operator) {
				updatedVal = validator

		bondedValidatorsCount++
		iterator.Next()

	// perform the actual kicks
	if oldCliffValidatorAddr != nil && kickCliffValidator {
		validator = getValidator(store, oldCliffValidatorAddr)
		unbondValidator(ctx, store, validator)
	return

// perform all the store operations for when a validator status becomes unbonded
unbondValidator(ctx Context, store KVStore, validator Validator)
	pool = GetPool(ctx)

	// set the status
	validator, pool = validator.UpdateStatus(pool, Unbonded)
	setPool(ctx, pool)

	// save the now unbonded validator record
	setValidator(validator)

	// add to accumulated changes for tendermint
	setTendermintUpdates(validator.abciValidatorZero)

	// also remove from the bonded validators index
	removeValidatorsBonded(validator)
}

// perform all the store operations for when a validator status becomes bonded
bondValidator(ctx Context, store KVStore, validator Validator) Validator
	pool = GetPool(ctx)

	// set the status
	validator, pool = validator.UpdateStatus(pool, Bonded)
	setPool(ctx, pool)

	// save the now bonded validator record to the three referenced stores
	setValidator(validator)
	setValidatorsBonded(validator)

	// add to accumulated changes for tendermint
	setTendermintUpdates(validator.abciValidator)

	return validator
```
