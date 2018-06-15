
### Transaction Overview

In this section we describe the processing of the transactions and the
corresponding updates to the state. Transactions: 
 - TxCreateValidator
 - TxEditValidator
 - TxDelegation
 - TxStartUnbonding
 - TxCompleteUnbonding
 - TxRedelegate
 - TxCompleteRedelegation

Other important state changes:
 - Update Validators

Other notes:
 - `tx` denotes a reference to the transaction being processed
 - `sender` denotes the address of the sender of the transaction
 - `getXxx`, `setXxx`, and `removeXxx` functions are used to retrieve and
    modify objects from the store
 - `sdk.Rat` refers to a rational numeric type specified by the SDK.
 
### TxCreateValidator

A validator is created using the `TxCreateValidator` transaction.

```golang
type TxCreateValidator struct {
	OwnerAddr           sdk.Address
    ConsensusPubKey     crypto.PubKey
    GovernancePubKey    crypto.PubKey
    SelfDelegation      coin.Coin       

    Description         Description
    Commission          sdk.Rat
    CommissionMax       sdk.Rat 
    CommissionMaxChange sdk.Rat 
}
	

createValidator(tx TxCreateValidator):
    validator = getValidator(tx.OwnerAddr)
    if validator != nil return // only one validator per address
   	
    validator = NewValidator(OwnerAddr, ConsensusPubKey, GovernancePubKey, Description)
    init validator poolShares, delegatorShares set to 0
    init validator commision fields from tx
    validator.PoolShares = 0
   	
    setValidator(validator)
   
    txDelegate = TxDelegate(tx.OwnerAddr, tx.OwnerAddr, tx.SelfDelegation) 
    delegate(txDelegate, validator) // see delegate function in [TxDelegate](TxDelegate)
    return
``` 

### TxEditValidator

If either the `Description` (excluding `DateBonded` which is constant),
`Commission`, or the `GovernancePubKey` need to be updated, the
`TxEditCandidacy` transaction should be sent from the owner account:

```golang
type TxEditCandidacy struct {
    GovernancePubKey    crypto.PubKey
    Commission          sdk.Rat
    Description         Description
}
 
editCandidacy(tx TxEditCandidacy):
    validator = getValidator(tx.ValidatorAddr)
    
    if tx.Commission > CommissionMax ||  tx.Commission < 0 then fail 
    if rateChange(tx.Commission) > CommissionMaxChange then fail
    validator.Commission = tx.Commission

    if tx.GovernancePubKey != nil validator.GovernancePubKey = tx.GovernancePubKey
    if tx.Description != nil validator.Description = tx.Description
    
    setValidator(store, validator)
    return
```
     	
### TxDelegation

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
    if validator.Status == Revoked return

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

		if bond.DelegatorAddr == validator.Owner && validator.Revoked == false 
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
		validator.Revoked = true

	validator = updateValidator(validator)

	if validator.DelegatorShares == 0 {
		removeValidator(validator.Owner)

    return
```

### TxCompleteUnbonding

Complete the unbonding and transfer the coins to the delegate. Perform any
slashing that occured during the unbonding period.

```golang
type TxUnbondingComplete struct {
    DelegatorAddr sdk.Address
    ValidatorAddr sdk.Address
}

redelegationComplete(tx TxRedelegate):
    unbonding = getUnbondingDelegation(tx.DelegatorAddr, tx.Validator)
    if unbonding.CompleteTime >= CurrentBlockTime && unbonding.CompleteHeight >= CurrentBlockHeight
        validator = GetValidator(tx.ValidatorAddr)
        returnTokens = ExpectedTokens * tx.startSlashRatio/validator.SlashRatio
	    AddCoins(unbonding.DelegatorAddr, returnTokens)
        removeUnbondingDelegation(unbonding)
    return     
```

### TxRedelegation

The redelegation command allows delegators to instantly switch validators. Once
the unbonding period has passed, the redelegation must be completed with
txRedelegationComplete.

```golang
type TxRedelegate struct {
    DelegatorAddr Address
    ValidatorFrom Validator
    ValidatorTo   Validator
    Shares        sdk.Rat 
    CompletedTime int64 
}

redelegate(tx TxRedelegate):

    pool = getPool()
    delegation = getDelegatorBond(tx.DelegatorAddr, tx.ValidatorFrom.Owner)
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

### TxCompleteRedelegation

Note that unlike TxCompleteUnbonding slashing of redelegating shares does not
take place during completion. Slashing on redelegated shares takes place
actively as a slashing occurs.

```golang
type TxRedelegationComplete struct {
    DelegatorAddr Address
    ValidatorFrom Validator
    ValidatorTo   Validator
}

redelegationComplete(tx TxRedelegate):
    redelegation = getRedelegation(tx.DelegatorAddr, tx.validatorFrom, tx.validatorTo)
    if redelegation.CompleteTime >= CurrentBlockTime && redelegation.CompleteHeight >= CurrentBlockHeight
        removeRedelegation(redelegation)
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

		ownerAddr = iterator.Value()
		if bytes.Equal(ownerAddr, newValidator.Owner) {
			validator = newValidator
        else
			validator = getValidator(ownerAddr)

		// if not previously a validator (and unrevoked),
		// kick the cliff validator / bond this new validator
		if validator.Status() != Bonded && !validator.Revoked {
			kickCliffValidator = true

			validator = bondValidator(ctx, store, validator)
			if bytes.Equal(ownerAddr, newValidator.Owner) {
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
