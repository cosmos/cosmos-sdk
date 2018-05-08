
### Transaction Overview

Available Transactions: 
* TxDeclareCandidacy
* TxEditCandidacy 
* TxDelegate
* TxUnbond 
* TxRedelegate
* TxLivelinessCheck
* TxProveLive

## Transaction processing

In this section we describe the processing of the transactions and the 
corresponding updates to the global state. In the following text we will use 
`gs` to refer to the `GlobalState` data structure, `unbondDelegationQueue` is a
reference to the queue of unbond delegations, `reDelegationQueue` is the 
reference for the queue of redelegations. We use `tx` to denote a 
reference to a transaction that is being processed, and `sender` to denote the 
address of the sender of the transaction. We use function 
`loadCandidate(store, PubKey)` to obtain a Candidate structure from the store, 
and `saveCandidate(store, candidate)` to save it. Similarly, we use 
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
    candidate = loadCandidate(store, tx.PubKey)
    if candidate != nil return // candidate with that public key already exists 
   	
    candidate = NewCandidate(tx.PubKey)
    candidate.Status = Unbonded
    candidate.Owner = sender
    init candidate VotingPower, GlobalStakeShares, IssuedDelegatorShares, RedelegatingShares and Adjustment to rational.Zero
    init commision related fields based on the values from tx
    candidate.ProposerRewardPool = Coin(0)  
    candidate.Description = tx.Description
   	
    saveCandidate(store, candidate)
   
    txDelegate = TxDelegate(tx.PubKey, tx.Amount)
    return delegateWithCandidate(txDelegate, candidate) 

// see delegateWithCandidate function in [TxDelegate](TxDelegate)
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
    candidate = loadCandidate(store, tx.PubKey)
    if candidate == nil or candidate.Status == Revoked return 
    
    if tx.GovernancePubKey != nil candidate.GovernancePubKey = tx.GovernancePubKey
    if tx.Commission >= 0 candidate.Commission = tx.Commission
    if tx.Description != nil candidate.Description = tx.Description
    
    saveCandidate(store, candidate)
    return
```
     	
### TxDelegate

Delegator bonds are created using the `TxDelegate` transaction. Within this 
transaction the delegator provides an amount of coins, and in return receives 
some amount of candidate's delegator shares that are assigned to 
`DelegatorBond.Shares`. 

```golang
type TxDelegate struct {
    PubKey crypto.PubKey
    Amount coin.Coin       
}

delegate(tx TxDelegate):
    candidate = loadCandidate(store, tx.PubKey)
    if candidate == nil return
	return delegateWithCandidate(tx, candidate)

delegateWithCandidate(tx TxDelegate, candidate Candidate):
    if candidate.Status == Revoked return

    if candidate.Status == Bonded 
	    poolAccount = params.HoldBonded
    else 
	    poolAccount = params.HoldUnbonded
	
    err = transfer(sender, poolAccount, tx.Amount)
    if err != nil return 

    bond = loadDelegatorBond(store, sender, tx.PubKey)
    if bond == nil then bond = DelegatorBond(tx.PubKey, rational.Zero, Coin(0), Coin(0))
	
    issuedDelegatorShares = addTokens(tx.Amount, candidate)
    bond.Shares += issuedDelegatorShares
	
    saveCandidate(store, candidate)
    saveDelegatorBond(store, sender, bond)
    saveGlobalState(store, gs)
    return 

addTokens(amount coin.Coin, candidate Candidate):
    if candidate.Status == Bonded 
	    gs.BondedPool += amount
	    issuedShares = amount / exchangeRate(gs.BondedShares, gs.BondedPool)
	    gs.BondedShares += issuedShares
    else 
	    gs.UnbondedPool += amount
	    issuedShares = amount / exchangeRate(gs.UnbondedShares, gs.UnbondedPool)
	    gs.UnbondedShares += issuedShares
	
    candidate.GlobalStakeShares += issuedShares
    
    if candidate.IssuedDelegatorShares.IsZero() 
        exRate = rational.One
    else
        exRate = candidate.GlobalStakeShares / candidate.IssuedDelegatorShares
	
    issuedDelegatorShares = issuedShares / exRate
    candidate.IssuedDelegatorShares += issuedDelegatorShares
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

    candidate = loadCandidate(store, tx.PubKey)
	
    revokeCandidacy = false
    if bond.Shares.IsZero() 
	    if sender == candidate.Owner and candidate.Status != Revoked then revokeCandidacy = true then removeDelegatorBond(store, sender, bond)
    else 
	    saveDelegatorBond(store, sender, bond)

    if candidate.Status == Bonded 
        poolAccount = params.HoldBonded
    else 
        poolAccount = params.HoldUnbonded

    returnedCoins = removeShares(candidate, shares)
	
    unbondDelegationElem = QueueElemUnbondDelegation(tx.PubKey, currentHeight(), sender, returnedCoins, startSlashRatio)
    unbondDelegationQueue.add(unbondDelegationElem)
	
    transfer(poolAccount, unbondingPoolAddress, returnCoins)  
    
    if revokeCandidacy 
	    if candidate.Status == Bonded then bondedToUnbondedPool(candidate)
	    candidate.Status = Revoked

    if candidate.IssuedDelegatorShares.IsZero() 
	    removeCandidate(store, tx.PubKey)
    else 
	    saveCandidate(store, candidate)

    saveGlobalState(store, gs)
    return 

removeShares(candidate Candidate, shares rational.Rat):
    globalPoolSharesToRemove = delegatorShareExRate(candidate) * shares

    if candidate.Status == Bonded 
	    gs.BondedShares -= globalPoolSharesToRemove
	    removedTokens = exchangeRate(gs.BondedShares, gs.BondedPool) * globalPoolSharesToRemove
	    gs.BondedPool -= removedTokens
    else 
	    gs.UnbondedShares -= globalPoolSharesToRemove
	    removedTokens = exchangeRate(gs.UnbondedShares, gs.UnbondedPool) * globalPoolSharesToRemove
	    gs.UnbondedPool -= removedTokens
	
    candidate.GlobalStakeShares -= removedTokens
    candidate.IssuedDelegatorShares -= shares
    return returnedCoins

delegatorShareExRate(candidate Candidate):
    if candidate.IssuedDelegatorShares.IsZero() then return rational.One
    return candidate.GlobalStakeShares / candidate.IssuedDelegatorShares
	
bondedToUnbondedPool(candidate Candidate):
    removedTokens = exchangeRate(gs.BondedShares, gs.BondedPool) * candidate.GlobalStakeShares 
    gs.BondedShares -= candidate.GlobalStakeShares
    gs.BondedPool -= removedTokens
	
    gs.UnbondedPool += removedTokens
    issuedShares = removedTokens / exchangeRate(gs.UnbondedShares, gs.UnbondedPool)
    gs.UnbondedShares += issuedShares
    
    candidate.GlobalStakeShares = issuedShares
    candidate.Status = Unbonded

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
    candidate = loadCandidate(store, tx.PubKeyFrom)
    if candidate == nil return
    
    candidate.RedelegatingShares += tx.Shares
    reDelegationElem = QueueElemReDelegate(tx.PubKeyFrom, currentHeight(), sender, tx.Shares, tx.PubKeyTo)
    redelegationQueue.add(reDelegationElem)
    return     
```

### TxLivelinessCheck

Liveliness issues are calculated by keeping track of the block precommits in
the block header. A queue is persisted which contains the block headers from
all recent blocks for the duration of the unbonding period. A validator is
defined as having livliness issues if they have not been included in more than
33% of the blocks over: 
* The most recent 24 Hours if they have >= 20% of global stake
* The most recent week if they have = 0% of global stake
* Linear interpolation of the above two scenarios

Liveliness kicks are only checked when a `TxLivelinessCheck` transaction is
submitted. 

```golang
type TxLivelinessCheck struct { 
    PubKey        crypto.PubKey
    RewardAccount Addresss
}
```

If the `TxLivelinessCheck` is successful in kicking a validator, 5% of the
liveliness punishment is provided as a reward to `RewardAccount`.

### TxProveLive

If the validator was kicked for liveliness issues and is able to regain
liveliness then all delegators in the temporary unbonding pool which have not
transacted to move will be bonded back to the now-live validator and begin to
once again collect provisions and rewards. Regaining liveliness is demonstrated
by sending in a `TxProveLive` transaction:

```golang
type TxProveLive struct {
    PubKey crypto.PubKey
}
```

