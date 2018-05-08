# Staking Module

## Overview

The Cosmos Hub is a Tendermint-based Proof of Stake blockchain system that 
serves as a backbone of the Cosmos ecosystem. It is operated and secured by an 
open and globally decentralized set of validators. Tendermint consensus is a 
Byzantine fault-tolerant distributed protocol that involves all validators in 
the process of exchanging protocol messages in the production of each block. To
avoid Nothing-at-Stake problem, a validator in Tendermint needs to lock up 
coins in a bond deposit. Tendermint protocol messages are signed by the 
validator's private key, and this is a basis for Tendermint strict 
accountability that allows punishing misbehaving validators by slashing 
(burning) their bonded Atoms. On the other hand, validators are rewarded for 
their service of securing blockchain network by the inflationary provisions and
transactions fees. This incentives correct behavior of the validators and 
provides the economic security of the network.

The native token of the Cosmos Hub is called Atom; becoming a validator of the 
Cosmos Hub requires holding Atoms. However, not all Atom holders are validators
of the Cosmos Hub. More precisely, there is a selection process that determines
the validator set as a subset of all validator candidates (Atom holders that 
wants to become a validator). The other option for Atom holder is to delegate 
their atoms to validators, i.e., being a delegator. A delegator is an Atom 
holder that has bonded its Atoms by delegating it to a validator (or validator 
candidate). By bonding Atoms to secure the network (and taking a risk of being 
slashed in case of misbehaviour), a user is rewarded with inflationary 
provisions and transaction fees proportional to the amount of its bonded Atoms.
The Cosmos Hub is designed to efficiently facilitate a small numbers of 
validators (hundreds), and large numbers of delegators (tens of thousands). 
More precisely, it is the role of the Staking module of the Cosmos Hub to
support various staking functionality including validator set selection, 
delegating, bonding and withdrawing Atoms, and the distribution of inflationary
provisions and transaction fees.

## State

The staking module persists the following information to the store:
* `GlobalState`, a struct describing the global pools, inflation, and
  fees
* `ValidatorCandidates: <pubkey | shares> => <candidate>`, a map of all candidates (including current validators) in the store,
indexed by their public key and shares in the global pool.
* `DelegatorBonds: < delegator-address | candidate-pubkey > => <delegator-bond>`. a map of all delegations by a delegator to a candidate,
indexed by delegator address and candidate pubkey.
  public key
* `UnbondQueue`, the queue of unbonding delegations
* `RedelegateQueue`, the queue of re-delegations

### Global State

The GlobalState contains information about the total amount of Atoms, the
global bonded/unbonded position, the Atom inflation rate, and the fees.

`Params` is global data structure that stores system parameters and defines overall functioning of the 
module.

``` go
type GlobalState struct {
    TotalSupply              int64        // total supply of Atoms
    BondedPool               int64        // reserve of bonded tokens
    BondedShares             rational.Rat // sum of all shares distributed for the BondedPool
    UnbondedPool             int64        // reserve of unbonding tokens held with candidates
    UnbondedShares           rational.Rat // sum of all shares distributed for the UnbondedPool
    InflationLastTime        int64        // timestamp of last processing of inflation
    Inflation                rational.Rat // current annual inflation rate
    DateLastCommissionReset  int64        // unix timestamp for last commission accounting reset
    FeePool                  coin.Coins   // fee pool for all the fee shares which have already been distributed
    ReservePool              coin.Coins   // pool of reserve taxes collected on all fees for governance use
    Adjustment               rational.Rat // Adjustment factor for calculating global fee accum
}

type Params struct {
    HoldBonded   Address // account  where all bonded coins are held
    HoldUnbonding Address // account where all delegated but unbonding coins are held

    InflationRateChange rational.Rational // maximum annual change in inflation rate
    InflationMax        rational.Rational // maximum inflation rate
    InflationMin        rational.Rational // minimum inflation rate
    GoalBonded          rational.Rational // Goal of percent bonded atoms
    ReserveTax          rational.Rational // Tax collected on all fees

    MaxVals          uint16  // maximum number of validators
    AllowedBondDenom string  // bondable coin denomination

    // gas costs for txs
    GasDeclareCandidacy int64 
    GasEditCandidacy    int64 
    GasDelegate         int64 
    GasRedelegate       int64 
    GasUnbond           int64 
}
```

### Candidate

The `Candidate` holds the current state and some historical 
actions of validators or candidate-validators. 

``` go
type Candidate struct {
    Status                 CandidateStatus       
    ConsensusPubKey        crypto.PubKey
    GovernancePubKey       crypto.PubKey
    Owner                  crypto.Address
    GlobalStakeShares      rational.Rat 
    IssuedDelegatorShares  rational.Rat
    RedelegatingShares     rational.Rat
    VotingPower            rational.Rat 
    Commission             rational.Rat
    CommissionMax          rational.Rat
    CommissionChangeRate   rational.Rat
    CommissionChangeToday  rational.Rat
    ProposerRewardPool     coin.Coins
    Adjustment             rational.Rat
    Description            Description 
}

type Description struct {
    Name       string 
    DateBonded string 
    Identity   string 
    Website    string 
    Details    string 
}
```

Candidate parameters are described:
* Status: it can be Bonded (active validator), Unbonding (validator candidate) 
  or Revoked
* ConsensusPubKey: candidate public key that is used strictly for participating in 
  consensus
* GovernancePubKey: public key used by the validator for governance voting 
* Owner: Address that is allowed to unbond coins.
* GlobalStakeShares: Represents shares of `GlobalState.BondedPool` if 
  `Candidate.Status` is `Bonded`; or shares of `GlobalState.Unbondingt Pool` 
  otherwise
* IssuedDelegatorShares: Sum of all shares a candidate issued to delegators 
  (which includes the candidate's self-bond); a delegator share represents 
  their stake in the Candidate's `GlobalStakeShares`
* RedelegatingShares: The portion of `IssuedDelegatorShares` which are 
  currently re-delegating to a new validator
* VotingPower: Proportional to the amount of bonded tokens which the validator
  has if `Candidate.Status` is `Bonded`; otherwise it is equal to `0`
* Commission:  The commission rate of fees charged to any delegators
* CommissionMax:  The maximum commission rate this candidate can charge each 
  day from the date `GlobalState.DateLastCommissionReset` 
* CommissionChangeRate: The maximum daily increase of the candidate commission
* CommissionChangeToday: Counter for the amount of change to commission rate 
  which has occurred today, reset on the first block of each day (UTC time)
* ProposerRewardPool: reward pool for extra fees collected when this candidate
  is the proposer of a block
* Adjustment factor used to passively calculate each validators entitled fees
  from `GlobalState.FeePool`
* Description
  * Name: moniker
  * DateBonded: date determined which the validator was bonded
  * Identity: optional field to provide a signature which verifies the 
    validators identity (ex. UPort or Keybase)
  * Website: optional website link
  * Details: optional details

### DelegatorBond

Atom holders may delegate coins to candidates; under this circumstance their
funds are held in a `DelegatorBond` data structure. It is owned by one 
delegator, and is associated with the shares for one candidate. The sender of 
the transaction is the owner of the bond.

``` go
type DelegatorBond struct {
    Candidate            crypto.PubKey
    Shares               rational.Rat
    AdjustmentFeePool    coin.Coins  
    AdjustmentRewardPool coin.Coins  
} 
```

Description: 
* Candidate: the public key of the validator candidate: bonding too
* Shares: the number of delegator shares received from the validator candidate
* AdjustmentFeePool: Adjustment factor used to passively calculate each bonds
  entitled fees from `GlobalState.FeePool`
* AdjustmentRewardPool: Adjustment factor used to passively calculate each
  bonds entitled fees from `Candidate.ProposerRewardPool`

 
### QueueElem

The Unbonding and re-delegation process is implemented using the ordered queue 
data structure. All queue elements share a common structure:

```golang
type QueueElem struct {
    Candidate   crypto.PubKey
    InitTime    int64    // when the element was added to the queue
}
```

The queue is ordered so the next element to unbond/re-delegate is at the head. 
Every tick the head of the queue is checked and if the unbonding period has 
passed since `InitTime`, the final settlement of the unbonding is started or 
re-delegation is executed, and the element is popped from the queue. Each 
`QueueElem` is persisted in the store until it is popped from the queue. 

### QueueElemUnbondDelegation

QueueElemUnbondDelegation structure is used in the unbonding queue. 

```golang
type QueueElemUnbondDelegation struct {
    QueueElem
    Payout           Address       // account to pay out to
    Tokens           coin.Coins    // the value in Atoms of the amount of delegator shares which are unbonding
    StartSlashRatio  rational.Rat  // candidate slash ratio 
}
``` 

### QueueElemReDelegate

QueueElemReDelegate structure is used in the re-delegation queue. 

```golang
type QueueElemReDelegate struct {
    QueueElem
    Payout       Address       // account to pay out to
    Shares       rational.Rat  // amount of shares which are unbonding
    NewCandidate crypto.PubKey // validator to bond to after unbond
}
```

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

### End of block handling

```golang
tick(ctx Context):
    hrsPerYr = 8766   // as defined by a julian year of 365.25 days
    
    time = ctx.Time()
    if time > gs.InflationLastTime + ProvisionTimeout 
        gs.InflationLastTime = time
        gs.Inflation = nextInflation(hrsPerYr).Round(1000000000)
        
        provisions = gs.Inflation * (gs.TotalSupply / hrsPerYr)
        
        gs.BondedPool += provisions
        gs.TotalSupply += provisions
        
        saveGlobalState(store, gs)
    
    if time > unbondDelegationQueue.head().InitTime + UnbondingPeriod 
        for each element elem in the unbondDelegationQueue where time > elem.InitTime + UnbondingPeriod do
    	    transfer(unbondingQueueAddress, elem.Payout, elem.Tokens)
    	    unbondDelegationQueue.remove(elem)
    
    if time > reDelegationQueue.head().InitTime + UnbondingPeriod 
        for each element elem in the unbondDelegationQueue where time > elem.InitTime + UnbondingPeriod do
            candidate = getCandidate(store, elem.PubKey)
            returnedCoins = removeShares(candidate, elem.Shares)
            candidate.RedelegatingShares -= elem.Shares 
            delegateWithCandidate(TxDelegate(elem.NewCandidate, returnedCoins), candidate)
            reDelegationQueue.remove(elem)
            
    return UpdateValidatorSet()

nextInflation(hrsPerYr rational.Rat):
    if gs.TotalSupply > 0 
        bondedRatio = gs.BondedPool / gs.TotalSupply
    else 
        bondedRation = 0
   
    inflationRateChangePerYear = (1 - bondedRatio / params.GoalBonded) * params.InflationRateChange
    inflationRateChange = inflationRateChangePerYear / hrsPerYr

    inflation = gs.Inflation + inflationRateChange
    if inflation > params.InflationMax then inflation = params.InflationMax
	
    if inflation < params.InflationMin then inflation = params.InflationMin
	
    return inflation 

UpdateValidatorSet():
    candidates = loadCandidates(store)

    v1 = candidates.Validators()
    v2 = updateVotingPower(candidates).Validators()

    change = v1.validatorsUpdated(v2) // determine all updated validators between two validator sets
    return change

updateVotingPower(candidates Candidates):
    foreach candidate in candidates do
	    candidate.VotingPower = (candidate.IssuedDelegatorShares - candidate.RedelegatingShares) * delegatorShareExRate(candidate)	
	    
    candidates.Sort()
	
    foreach candidate in candidates do
	    if candidate is not in the first params.MaxVals  
	        candidate.VotingPower = rational.Zero
	        if candidate.Status == Bonded then bondedToUnbondedPool(candidate Candidate)
		
	    else if candidate.Status == UnBonded then unbondedToBondedPool(candidate)
                      
	saveCandidate(store, c)
	
    return candidates

unbondedToBondedPool(candidate Candidate):
    removedTokens = exchangeRate(gs.UnbondedShares, gs.UnbondedPool) * candidate.GlobalStakeShares 
    gs.UnbondedShares -= candidate.GlobalStakeShares
    gs.UnbondedPool -= removedTokens
	
    gs.BondedPool += removedTokens
    issuedShares = removedTokens / exchangeRate(gs.BondedShares, gs.BondedPool)
    gs.BondedShares += issuedShares
    
    candidate.GlobalStakeShares = issuedShares
    candidate.Status = Bonded

    return transfer(address of the unbonded pool, address of the bonded pool, removedTokens)
```
