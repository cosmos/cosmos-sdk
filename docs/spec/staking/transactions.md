
### Transaction Overview

In this section we describe the processing of the transactions and the
corresponding updates to the state. Transactions: 
 - TxCreateValidator
 - TxEditValidator
 - TxDelegation
 - TxRedelegation
 - TxUnbond 

Other important state changes:
 - Update Validators

Other notes:
 - `tx` denotes a reference to the transaction being processed
 - `sender` denotes the address of the sender of the transaction
 - `getXxx`, `setXxx`, and `removeXxx` functions are used to retrieve and
    modify objects from the store
 - `sdk.Rat` refers to a rational numeric type specified by the sdk.
 
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
    init validator poolShares, delegatorShares set to 0 //XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
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
    
    if tx.Commission > CommissionMax ||  tx.Commission < 0 return halt tx
    if rateChange(tx.Commission) > CommissionMaxChange return halt tx
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

### TxUnbond 

Delegator unbonding is defined with the following transaction:

```golang
type TxUnbond struct {
	DelegatorAddr sdk.Address 
	ValidatorAddr sdk.Address 
	Shares        string      
}

unbond(tx TxUnbond):    
    delegation, found = getDelegatorBond(store, sender, tx.PubKey)
    if !found == nil return 
    
	if msg.Shares == "MAX" {
		if !bond.Shares.GT(sdk.ZeroRat()) {
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()
    else 
		var err sdk.Error
		delShares, err = sdk.NewRatFromDecimal(msg.Shares)
		if err != nil 
            return err
		if bond.Shares.LT(delShares) 
			return ErrNotEnoughBondShares(k.codespace, msg.Shares).Result()

	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return err 

	if msg.Shares == "MAX" 
		delShares = bond.Shares

	bond.Shares -= delShares
    
    unbondingDelegation = NewUnbondingDelegation(sender, delShares, currentHeight/Time, startSlashRatio)
    setUnbondingDelegation(unbondingDelegation)

	revokeCandidacy := false
	if bond.Shares.IsZero() {

		if bond.DelegatorAddr == validator.Owner && validator.Revoked == false 
			revokeCandidacy = true

		k.removeDelegation(ctx, bond)
	else
		bond.Height = currentBlockHeight
		setDelegation(bond)

	pool := k.GetPool(ctx)
	validator, pool, returnAmount := validator.removeDelShares(pool, delShares)
	k.setPool(ctx, pool)
	AddCoins(ctx, bond.DelegatorAddr, returnAmount)

	if revokeCandidacy
		validator.Revoked = true

	validator = updateValidator(ctx, validator)

	if validator.DelegatorShares == 0 {
		removeValidator(ctx, validator.Owner)

    return
```

### TxRedelegation

The redelegation command allows delegators to instantly switch validators. 

```golang
type TxRedelegate struct {
    DelegatorAddr Address
    ValidatorFrom Validator
    ValidatorTo   Validator
    Shares        sdk.Rat 
}

redelegate(tx TxRedelegate):
    pool = getPool()
    delegation = getDelegatorBond(tx.DelegatorAddr, tx.ValidatorFrom.Owner)
    if delegation == nil then return 
    
    if delegation.Shares < tx.Shares return 
    delegation.shares -= Tx.Shares
    validator, pool, createdCoins = validator.RemoveShares(pool, tx.Shares)
    setPool(pool)
    
    redelegation = newRedelegation(validatorFrom, validatorTo, Shares, createdCoins)
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

	kickCliffValidator := false
	oldCliffValidatorAddr := getCliffValidator(ctx)

	// add the actual validator power sorted store
	maxValidators := GetParams(ctx).MaxValidators
	iterator := ReverseSubspaceIterator(ValidatorsByPowerKey) // largest to smallest
	bondedValidatorsCount := 0
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

		ownerAddr := iterator.Value()
		if bytes.Equal(ownerAddr, newValidator.Owner) {
			validator = newValidator
        else
			validator = getValidator(ownerAddr)

		// if not previously a validator (and unrevoked),
		// kick the cliff validator / bond this new validator
		if validator.Status() != sdk.Bonded && !validator.Revoked {
			kickCliffValidator = true

			validator = bondValidator(ctx, store, validator)
			if bytes.Equal(ownerAddr, newValidator.Owner) {
				updatedVal = validator

		bondedValidatorsCount++
		iterator.Next()

	// perform the actual kicks
	if oldCliffValidatorAddr != nil && kickCliffValidator {
		validator := getValidator(store, oldCliffValidatorAddr)
		unbondValidator(ctx, store, validator)
	return

// perform all the store operations for when a validator status becomes unbonded
unbondValidator(ctx sdk.Context, store sdk.KVStore, validator Validator)
	pool := GetPool(ctx)

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	setPool(ctx, pool)

	// save the now unbonded validator record
	setValidator(validator)

	// add to accumulated changes for tendermint
	setTendermintUpdates(validator.abciValidatorZero)

	// also remove from the bonded validators index
	removeValidatorsBonded(validator)
}

// perform all the store operations for when a validator status becomes bonded
bondValidator(ctx sdk.Context, store sdk.KVStore, validator Validator) Validator 
	pool := GetPool(ctx)

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	setPool(ctx, pool)

	// save the now bonded validator record to the three referenced stores
	setValidator(validator)
	setValidatorsBonded(validator)

	// add to accumulated changes for tendermint
	setTendermintUpdates(validator.abciValidator)

	return validator
```
