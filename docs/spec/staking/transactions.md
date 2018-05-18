
### Transaction Overview

Available Transactions: 
* TxDeclareCandidacy
* TxEditCandidacy 
* TxDelegate
* TxUnbond 
* TxRedelegate
* TxProveLive

## Transaction processing

In this section we describe the processing of the transactions and the 
corresponding updates to the global state. In the following text we will use 
`gs` to refer to the `GlobalState` data structure, `unbondDelegationQueue` is a
reference to the queue of unbond delegations, `reDelegationQueue` is the 
reference for the queue of redelegations. We use `tx` to denote a 
reference to a transaction that is being processed, and `sender` to denote the 
address of the sender of the transaction. We use function 
`loadValidator(store, PubKey)` to obtain a Validator structure from the store, 
and `saveValidator(store, validator)` to save it. Similarly, we use 
`loadDelegatorBond(store, sender, PubKey)` to load a delegator bond with the 
key (sender and PubKey) from the store, and 
`saveDelegatorBond(store, sender, bond)` to save it. 
`removeDelegatorBond(store, sender, bond)` is used to remove the bond from the 
store.
 
### TxDeclareCandidacy

A validator candidacy is declared using the `TxDeclareCandidacy` transaction.

```golang
type TxDeclareCandidacy struct {
    ConsensusPubKey     crypto.PubKey
    Amount              coin.Coin       
    GovernancePubKey    crypto.PubKey
    Commission          rational.Rat
    CommissionMax       int64 
    CommissionMaxChange int64 
    Description         Description
}

declareCandidacy(tx TxDeclareCandidacy):
    validator = loadValidator(store, tx.PubKey)
    if validator != nil return // validator with that public key already exists 
   	
    validator = NewValidator(tx.PubKey)
    validator.Status = Unbonded
    validator.Owner = sender
    init validator VotingPower, GlobalStakeShares, IssuedDelegatorShares, RedelegatingShares and Adjustment to rational.Zero
    init commision related fields based on the values from tx
    validator.ProposerRewardPool = Coin(0)  
    validator.Description = tx.Description
   	
    saveValidator(store, validator)
   
    txDelegate = TxDelegate(tx.PubKey, tx.Amount)
    return delegateWithValidator(txDelegate, validator) 

// see delegateWithValidator function in [TxDelegate](TxDelegate)
``` 

### TxEditCandidacy

If either the `Description` (excluding `DateBonded` which is constant),
`Commission`, or the `GovernancePubKey` need to be updated, the
`TxEditCandidacy` transaction should be sent from the owner account:

```golang
type TxEditCandidacy struct {
    GovernancePubKey    crypto.PubKey
    Commission          int64  
    Description         Description
}
 
editCandidacy(tx TxEditCandidacy):
    validator = loadValidator(store, tx.PubKey)
    if validator == nil or validator.Status == Revoked return 
    
    if tx.GovernancePubKey != nil validator.GovernancePubKey = tx.GovernancePubKey
    if tx.Commission >= 0 validator.Commission = tx.Commission
    if tx.Description != nil validator.Description = tx.Description
    
    saveValidator(store, validator)
    return
```
     	
### TxDelegate

Delegator bonds are created using the `TxDelegate` transaction. Within this 
transaction the delegator provides an amount of coins, and in return receives 
some amount of validator's delegator shares that are assigned to 
`DelegatorBond.Shares`. 

```golang
type TxDelegate struct {
    PubKey crypto.PubKey
    Amount coin.Coin       
}

delegate(tx TxDelegate):
    validator = loadValidator(store, tx.PubKey)
    if validator == nil return
	return delegateWithValidator(tx, validator)

delegateWithValidator(tx TxDelegate, validator Validator):
    if validator.Status == Revoked return

    if validator.Status == Bonded 
	    poolAccount = params.HoldBonded
    else 
	    poolAccount = params.HoldUnbonded
	
    err = transfer(sender, poolAccount, tx.Amount)
    if err != nil return 

    bond = loadDelegatorBond(store, sender, tx.PubKey)
    if bond == nil then bond = DelegatorBond(tx.PubKey, rational.Zero, Coin(0), Coin(0))
	
    issuedDelegatorShares = addTokens(tx.Amount, validator)
    bond.Shares += issuedDelegatorShares
	
    saveValidator(store, validator)
    saveDelegatorBond(store, sender, bond)
    saveGlobalState(store, gs)
    return 

addTokens(amount coin.Coin, validator Validator):
    if validator.Status == Bonded 
	    gs.BondedPool += amount
	    issuedShares = amount / exchangeRate(gs.BondedShares, gs.BondedPool)
	    gs.BondedShares += issuedShares
    else 
	    gs.UnbondedPool += amount
	    issuedShares = amount / exchangeRate(gs.UnbondedShares, gs.UnbondedPool)
	    gs.UnbondedShares += issuedShares
	
    validator.GlobalStakeShares += issuedShares
    
    if validator.IssuedDelegatorShares.IsZero() 
        exRate = rational.One
    else
        exRate = validator.GlobalStakeShares / validator.IssuedDelegatorShares
	
    issuedDelegatorShares = issuedShares / exRate
    validator.IssuedDelegatorShares += issuedDelegatorShares
    return issuedDelegatorShares
	
exchangeRate(shares rational.Rat, tokenAmount int64):
    if shares.IsZero() then return rational.One
    return tokenAmount / shares
    	
```

### TxUnbond 

Delegator unbonding is defined with the following transaction:

```golang
type TxUnbond struct {
    PubKey crypto.PubKey
    Shares rational.Rat 
}

unbond(tx TxUnbond):    
    bond = loadDelegatorBond(store, sender, tx.PubKey)
    if bond == nil return 
    if bond.Shares < tx.Shares return 
	
    bond.Shares -= tx.Shares

    validator = loadValidator(store, tx.PubKey)
	
    revokeCandidacy = false
    if bond.Shares.IsZero() 
	    if sender == validator.Owner and validator.Status != Revoked then revokeCandidacy = true then removeDelegatorBond(store, sender, bond)
    else 
	    saveDelegatorBond(store, sender, bond)

    if validator.Status == Bonded 
        poolAccount = params.HoldBonded
    else 
        poolAccount = params.HoldUnbonded

    returnedCoins = removeShares(validator, shares)
	
    unbondDelegationElem = QueueElemUnbondDelegation(tx.PubKey, currentHeight(), sender, returnedCoins, startSlashRatio)
    unbondDelegationQueue.add(unbondDelegationElem)
	
    transfer(poolAccount, unbondingPoolAddress, returnCoins)  
    
    if revokeCandidacy 
	    if validator.Status == Bonded then bondedToUnbondedPool(validator)
	    validator.Status = Revoked

    if validator.IssuedDelegatorShares.IsZero() 
	    removeValidator(store, tx.PubKey)
    else 
	    saveValidator(store, validator)

    saveGlobalState(store, gs)
    return 

removeShares(validator Validator, shares rational.Rat):
    globalPoolSharesToRemove = delegatorShareExRate(validator) * shares

    if validator.Status == Bonded 
	    gs.BondedShares -= globalPoolSharesToRemove
	    removedTokens = exchangeRate(gs.BondedShares, gs.BondedPool) * globalPoolSharesToRemove
	    gs.BondedPool -= removedTokens
    else 
	    gs.UnbondedShares -= globalPoolSharesToRemove
	    removedTokens = exchangeRate(gs.UnbondedShares, gs.UnbondedPool) * globalPoolSharesToRemove
	    gs.UnbondedPool -= removedTokens
	
    validator.GlobalStakeShares -= removedTokens
    validator.IssuedDelegatorShares -= shares
    return returnedCoins

delegatorShareExRate(validator Validator):
    if validator.IssuedDelegatorShares.IsZero() then return rational.One
    return validator.GlobalStakeShares / validator.IssuedDelegatorShares
	
bondedToUnbondedPool(validator Validator):
    removedTokens = exchangeRate(gs.BondedShares, gs.BondedPool) * validator.GlobalStakeShares 
    gs.BondedShares -= validator.GlobalStakeShares
    gs.BondedPool -= removedTokens
	
    gs.UnbondedPool += removedTokens
    issuedShares = removedTokens / exchangeRate(gs.UnbondedShares, gs.UnbondedPool)
    gs.UnbondedShares += issuedShares
    
    validator.GlobalStakeShares = issuedShares
    validator.Status = Unbonded

    return transfer(address of the bonded pool, address of the unbonded pool, removedTokens)
```

### TxRedelegate

The re-delegation command allows delegators to switch validators while still
receiving equal reward to as if they had never unbonded.

```golang
type TxRedelegate struct {
    PubKeyFrom crypto.PubKey
    PubKeyTo   crypto.PubKey
    Shares     rational.Rat 
}

redelegate(tx TxRedelegate):
    bond = loadDelegatorBond(store, sender, tx.PubKey)
    if bond == nil then return 
    
    if bond.Shares < tx.Shares return 
    validator = loadValidator(store, tx.PubKeyFrom)
    if validator == nil return
    
    validator.RedelegatingShares += tx.Shares
    reDelegationElem = QueueElemReDelegate(tx.PubKeyFrom, currentHeight(), sender, tx.Shares, tx.PubKeyTo)
    redelegationQueue.add(reDelegationElem)
    return     
```

### TxProveLive

If a validator was automatically unbonded due to liveness issues and wishes to
assert it is still online, it can send `TxProveLive`:

```golang
type TxProveLive struct {
    PubKey crypto.PubKey
}
```

All delegators in the temporary unbonding pool which have not
transacted to move will be bonded back to the now-live validator and begin to
once again collect provisions and rewards. 

```
TODO: pseudo-code
```
