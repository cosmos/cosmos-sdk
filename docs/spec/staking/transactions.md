
### Transaction Overview

In this section we describe the processing of the transactions and the
corresponding updates to the state. 

Available Transactions: 
 - TxCreateValidator
 - TxEditValidator
 - TxDelegation
 - TxRedelegation
 - TxUnbond 

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

