# Staking Module

## Overview

The Cosmos Hub is a Tendermint-based Proof of Stake blockchain system that serves as a backbone of the Cosmos ecosystem.
It is operated and secured by an open and globally decentralized set of validators. Tendermint consensus is a 
Byzantine fault-tolerant distributed protocol that involves all validators in the process of exchanging protocol 
messages in the production of each block. To avoid Nothing-at-Stake problem, a validator in Tendermint needs to lock up
coins in a bond deposit. Tendermint protocol messages are signed by the validator's private key, and this is a basis for 
Tendermint strict accountability that allows punishing misbehaving validators by slashing (burning) their bonded Atoms. 
On the other hand, validators are for it's service of securing blockchain network rewarded by the inflationary 
provisions and transactions fees. This incentivizes correct behavior of the validators and provide economic security 
of the network.

The native token of the Cosmos Hub is called Atom; becoming a validator of the Cosmos Hub requires holding Atoms. 
However, not all Atom holders are validators of the Cosmos Hub. More precisely, there is a selection process that 
determines the validator set as a subset of all validator candidates (Atom holder that wants to 
become a validator). The other option for Atom holder is to delegate their atoms to validators, i.e., 
being a delegator. A delegator is an Atom holder that has bonded its Atoms by delegating it to a validator 
(or validator candidate). By bonding Atoms to securing network (and taking a risk of being slashed in case the 
validator misbehaves), a user is rewarded with inflationary provisions and transaction fees proportional to the amount 
of its bonded Atoms. The Cosmos Hub is designed to efficiently facilitate a small numbers of validators (hundreds), and 
large numbers of delegators (tens of thousands). More precisely, it is the role of the Staking module of the Cosmos Hub 
to support various staking functionality including validator set selection; delegating, bonding and withdrawing Atoms; 
and the distribution of inflationary provisions and transaction fees.
  
## State

The staking module persists the following information to the store:
- `GlobalState`, describing the global pools and the inflation related fields
- `map[PubKey]Candidate`, a map of validator candidates (including current validators), indexed by public key
- `map[rational.Rat]Candidate`, an ordered map of validator candidates (including current validators), indexed by 
shares in the global pool (bonded or unbonded depending on candidate status) 
- `map[[]byte]DelegatorBond`, a map of DelegatorBonds (for each delegation to a candidate by a delegator), indexed by 
the delegator address and the candidate public key
- `queue[QueueElemUnbondDelegation]`, a queue of unbonding delegations
- `queue[QueueElemReDelegate]`, a queue of re-delegations

### Global State

GlobalState data structure contains total Atoms supply, amount of Atoms in the bonded pool, sum of all shares 
distributed for the bonded pool, amount of Atoms in the unbonded pool, sum of all shares distributed for the 
unbonded pool, a timestamp of the last processing of inflation, the current annual inflation rate, a timestamp 
for the last comission accounting reset, the global fee pool, a pool of reserve taxes collected for the governance use
and an adjustment factor for calculating global feel accum (?).    
  
``` golang
type GlobalState struct {
    TotalSupply              int64        // total supply of Atoms
    BondedPool               int64        // reserve of bonded tokens
    BondedShares             rational.Rat // sum of all shares distributed for the BondedPool
    UnbondedPool             int64        // reserve of unbonded tokens held with candidates
    UnbondedShares           rational.Rat // sum of all shares distributed for the UnbondedPool
    InflationLastTime        int64        // timestamp of last processing of inflation
    Inflation                rational.Rat // current annual inflation rate
    DateLastCommissionReset  int64        // unix timestamp for last commission accounting reset
    FeePool                  coin.Coins   // fee pool for all the fee shares which have already been distributed
    ReservePool              coin.Coins   // pool of reserve taxes collected on all fees for governance use
    Adjustment               rational.Rat // Adjustment factor for calculating global fee accum
}
```

### Candidate

The `Candidate` data structure holds the current state and some historical actions of
validators or candidate-validators. 

``` golang
type Candidate struct {
    Status                 CandidateStatus       
    PubKey                 crypto.PubKey
    GovernancePubKey       crypto.PubKey
    Owner                  Address
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
```

CandidateStatus can be VyingUnbonded, VyingUnbonding, Bonded, KickUnbonding and KickUnbonded.


``` golang
type Description struct {
	Name       string 
	DateBonded string 
	Identity   string 
	Website    string 
	Details    string 
}
```

Candidate parameters are described:
 - Status: signal that the candidate is either vying for validator status,
   either unbonded or unbonding, an active validator, or a kicked validator
   either unbonding or unbonded.
 - PubKey: separated key from the owner of the candidate as is used strictly
   for participating in consensus.
 - Owner: Address where coins are bonded from and unbonded to 
 - GlobalStakeShares: Represents shares of `GlobalState.BondedPool` if
   `Candidate.Status` is `Bonded`; or shares of `GlobalState.UnbondedPool` otherwise
 - IssuedDelegatorShares: Sum of all shares a candidate issued to delegators (which
   includes the candidate's self-bond); a delegator share represents their stake in
   the Candidate's `GlobalStakeShares`
 - RedelegatingShares: The portion of `IssuedDelegatorShares` which are
   currently re-delegating to a new validator
 - VotingPower: Proportional to the amount of bonded tokens which the validator
   has if the candidate is a validator.
 - Commission:  The commission rate of fees charged to any delegators
 - CommissionMax:  The maximum commission rate this candidate can charge 
   each day from the date `GlobalState.DateLastCommissionReset` 
 - CommissionChangeRate: The maximum daily increase of the candidate commission
 - CommissionChangeToday: Counter for the amount of change to commission rate 
   which has occurred today, reset on the first block of each day (UTC time)
 - ProposerRewardPool: reward pool for extra fees collected when this candidate
   is the proposer of a block
 - Adjustment factor used to passively calculate each validators entitled fees
   from `GlobalState.FeePool`
 - Description
   - Name: moniker
   - DateBonded: date determined which the validator was bonded
   - Identity: optional field to provide a signature which verifies the
     validators identity (ex. UPort or Keybase)
   - Website: optional website link
   - Details: optional details

### DelegatorBond

Atom holders may delegate coins to validators; under this circumstance their
funds are held in a `DelegatorBond` data structure. It is owned by one delegator, and is
associated with the shares for one validator. The sender of the transaction is
considered the owner of the bond.  

``` golang
type DelegatorBond struct {
	Candidate            crypto.PubKey
	Shares               rational.Rat
    AdjustmentFeePool    coin.Coins  
    AdjustmentRewardPool coin.Coins  
} 
```

Description: 
 - Candidate: the public key of the validator candidate: bonding too
 - Shares: the number of delegator shares received from the validator candidate
 - AdjustmentFeePool: Adjustment factor used to passively calculate each bonds
   entitled fees from `GlobalState.FeePool`
 - AdjustmentRewardPool: Adjustment factor used to passively calculate each
   bonds entitled fees from `Candidate.ProposerRewardPool``
 
### QueueElem

Unbonding and re-delegation process is implemented using the ordered queue data structure. 
All queue elements used share a common structure:

``` golang
type QueueElem struct {
	Candidate   crypto.PubKey
	InitHeight  int64    // when the queue was initiated
}
```

The queue is ordered so the next to unbond/re-delegate is at the head. Every
tick the head of the queue is checked and if the unbonding period has passed
since `InitHeight`, the final settlement of the unbonding is started or re-delegation is executed, and the element is
pop from the queue. Each `QueueElem` is persisted in the store until it is popped from the queue. 

### QueueElemUnbondDelegation

``` golang
type QueueElemUnbondDelegation struct {
	QueueElem
	Payout           Address  // account to pay out to
    Shares           rational.Rat  // amount of delegator shares which are unbonding
    StartSlashRatio  rational.Rat  // candidate slash ratio at start of re-delegation
}
``` 
In the unbonding queue - the fraction of all historical slashings on
that validator are recorded (`StartSlashRatio`). When this queue reaches maturity
if that total slashing applied is greater on the validator then the
difference (amount that should have been slashed from the first validator) is
assigned to the amount being paid out.  

### QueueElemReDelegate

``` golang
type QueueElemReDelegate struct {
	QueueElem
	Payout       Address  // account to pay out to
    Shares       rational.Rat  // amount of shares which are unbonding
    NewCandidate crypto.PubKey // validator to bond to after unbond
}
```

### Transaction Overview

Available Transactions: 
 - TxDeclareCandidacy
 - TxEditCandidacy
 - TxLivelinessCheck
 - TxProveLive 
 - TxDelegate
 - TxUnbond 
 - TxRedelegate

## Transaction processing

In this section we describe the processing of the transactions and the corresponding updates to the global state.
For the following text we will use gs to refer to the GlobalState data structure, candidateMap is a reference to the 
map[PubKey]Candidate, delegatorBonds is a reference to map[[]byte]DelegatorBond, unbondDelegationQueue is a 
reference to the queue[QueueElemUnbondDelegation] and redelegationQueue is the reference for the 
queue[QueueElemReDelegate]. We use tx to denote reference to a transaction that is being processed.      
 
### TxDeclareCandidacy

A validator candidacy can be declared using the `TxDeclareCandidacy` transaction.
During this transaction a self-delegation transaction is executed to bond
tokens which are sent in with the transaction (TODO: What does this mean?).

``` golang
type TxDeclareCandidacy struct {
    PubKey              crypto.PubKey
    Amount              coin.Coin       
    GovernancePubKey    crypto.PubKey
    Commission          rational.Rat
    CommissionMax       int64 
    CommissionMaxChange int64 
    Description         Description
}
``` 

``` 
declareCandidacy(tx TxDeclareCandidacy):
    // create and save the empty candidate
    candidate = loadCandidate(store, tx.PubKey)
    if candidate != nil then return 
   	
    candidate = NewCandidate(tx.PubKey)
    candidate.Status = Unbonded
    candidate.Owner = sender
    init candidate VotingPower, GlobalStakeShares, IssuedDelegatorShares,RedelegatingShares and Adjustment to rational.Zero
    init commision related fields based on the values from tx
    candidate.ProposerRewardPool = Coin(0)  
    candidate.Description = tx.Description
   	
    saveCandidate(store, candidate)
   
    // move coins from the sender account to a (self-bond) delegator account
    // the candidate account and global shares are updated within here
    txDelegate = TxDelegate{tx.BondUpdate}
    return delegateWithCandidate(txDelegate, candidate)
``` 

### TxEditCandidacy

If either the `Description` (excluding `DateBonded` which is constant),
`Commission`, or the `GovernancePubKey` need to be updated, the
`TxEditCandidacy` transaction should be sent from the owner account:

``` golang
type TxEditCandidacy struct {
    GovernancePubKey    crypto.PubKey
    Commission          int64  
    Description         Description
}
```
 
```
editCandidacy(tx TxEditCandidacy):
    candidate = loadCandidate(store, tx.PubKey)
    if candidate == nil or candidate.Status == Unbonded return 
    if tx.GovernancePubKey != nil then candidate.GovernancePubKey = tx.GovernancePubKey
    if tx.Commission >= 0 then candidate.Commission = tx.Commission
    if tx.Description != nil then candidate.Description = tx.Description
    saveCandidate(store, candidate)
    return
  ```
     	
### TxDelegate

All bonding, whether self-bonding or delegation, is done via `TxDelegate`. 

Delegator bonds are created using the `TxDelegate` transaction. Within this transaction the delegator provides 
an amount of coins, and in return receives some amount of candidate's delegator shares that are assigned to 
`DelegatorBond.Shares`. The amount of created delegator shares depends on the candidate's 
delegator-shares-to-atoms exchange rate and is computed as
`delegator-shares = delegator-coins / delegator-shares-to-atom-ex-rate`.

``` golang 
type TxDelegate struct { 
	PubKey crypto.PubKey
	Amount coin.Coin       
}
```

```
delegate(tx TxDelegate):
    candidate = loadCandidate(store, tx.PubKey)
    if candidate == nil then return
	return delegateWithCandidate(tx, candidate)

delegateWithCandidate(tx TxDelegate, candidate Candidate):
    if candidate.Status == Revoked then return

	if candidate.Status == Bonded then
		poolAccount = address of the bonded pool
	else 
		poolAccount = address of the unbonded pool
	
	// Move coins from the delegator account to the bonded pool account
	err = transfer(sender, poolAccount, tx.Amount)
	if err != nil then return 

	// Get or create the delegator bond
	bond = loadDelegatorBond(store, sender, tx.PubKey)
	if bond == nil then 
	    bond = DelegatorBond{tx.PubKey,rational.Zero, Coin(0), Coin(0)}
	
	issuedDelegatorShares = candidate.addTokens(tx.Amount, gs)
	bond.Shares = bond.Shares.Add(issuedDelegatorShares)
	
	saveCandidate(store, candidate)
	
	store.Set(GetDelegatorBondKey(sender, bond.PubKey), bond)
	
	saveGlobalState(store, gs)
	return 

addTokens(amount int64, gs GlobalState, candidate Candidate):

	// get the exchange rate of global pool shares over delegator shares
    if candidate.IssuedDelegatorShares.IsZero() then 
        exRate = rational.One
    else
        exRate = candiate.GlobalStakeShares.Quo(candidate.IssuedDelegatorShares)
    
	if candidate.Status == Bonded then
		gs.BondedPool += amount
		issuedShares = exchangeRate(gs.BondedShares, gs.BondedPool).Inv().Mul(amount) // (tokens/shares)^-1 * tokens
        gs.BondedShares = gs.BondedShares.Add(issuedShares)
	else 
		gs.UnbondedPool += amount
		issuedShares = exchangeRate(gs.UnbondedShares, gs.UnbondedPool).Inv().Mul(amount) // (tokens/shares)^-1 * tokens
        gs.UnbondedShares = gs.UnbondedShares.Add(issuedShares)
	
	candidate.GlobalStakeShares = candidate.GlobalStakeShares.Add(issuedShares)

	issuedDelegatorShares = exRate.Mul(receivedGlobalShares)
	candidate.IssuedDelegatorShares = candidate.IssuedDelegatorShares.Add(issuedDelegatorShares)
	return
	
exchangeRate(shares rational.Rat, tokenAmount int64):
    if shares.IsZero() then return rational.One
    return shares.Inv().Mul(tokenAmount)
    	
```

### TxUnbond 
Delegator unbonding is defined with the following transaction:

``` golang
type TxUnbond struct { 
	PubKey crypto.PubKey
	Shares rational.Rat 
}
```

```
unbond(tx TxUnbond):

	// get delegator bond
	bond = loadDelegatorBond(store, sender, tx.PubKey)
	if bond == nil then return 

	// subtract bond tokens from delegator bond
	if bond.Shares.LT(tx.Shares) return // bond shares < tx shares
	
	bond.Shares = bond.Shares.Sub(ts.Shares)

	candidate = loadCandidate(store, tx.PubKey)
	if candidate == nil return

	revokeCandidacy = false
	if bond.Shares.IsZero() {
        // if the bond is the owner of the candidate then trigger a revoke candidacy
		if sender.Equals(candidate.Owner) and candidate.Status != Revoked then
			revokeCandidacy = true

		// remove the bond
		removeDelegatorBond(store, sender, tx.PubKey)
	else 
	    saveDelegatorBond(store, sender, bond)

	// transfer coins back to account
	if candidate.Status == Bonded then
        poolAccount = address of the bonded pool
    else 
        poolAccount = address of the unbonded pool

	returnCoins = candidate.removeShares(shares, gs)
	// TODO: Shouldn't it be created a queue element in this case?
	transfer(poolAccount, sender, returnCoins)

	if revokeCandidacy then
	    // change the share types to unbonded if they were not already
	if candidate.Status == Bonded then
	    // replace bonded shares with unbonded shares
        tokens = gs.removeSharesBonded(candidate.GlobalStakeShares)
        candidate.GlobalStakeShares = gs.addTokensUnbonded(tokens)
        candidate.Status = Unbonded
        
        transfer(address of the bonded pool, address of the unbonded pool, tokens)
		// lastly update the status
		candidate.Status = Revoked

	// deduct shares from the candidate and save
	if candidate.GlobalStakeShares.IsZero() then
		removeCandidate(store, tx.PubKey)
	else 
		saveCandidate(store, candidate)

	saveGlobalState(store, gs)
	return 
	
removeDelegatorBond(candidate Candidate):

	// first remove from the list of bonds
	pks = loadDelegatorCandidates(store, sender)
	for i, pk := range pks {
		if candidate.Equals(pk) {
			pks = append(pks[:i], pks[i+1:]...)
		}
	}
	b := wire.BinaryBytes(pks)
	store.Set(GetDelegatorBondsKey(delegator), b)

	// now remove the actual bond
	store.Remove(GetDelegatorBondKey(delegator, candidate))
	//updateDelegatorBonds(store, delegator)
}	
```

### Inflation provisions

Validator provisions are minted on an hourly basis (the first block of a new
hour).  The annual target of between 7% and 20%. The long-term target ratio of
bonded tokens to unbonded tokens is 67%.  
    
The target annual inflation rate is recalculated for each previsions cycle. The
inflation is also subject to a rate change (positive of negative) depending or
the distance from the desired ratio (67%). The maximum rate change possible is
defined to be 13% per year, however the annual inflation is capped as between
7% and 20%.
    
```
inflationRateChange(0) = 0
GlobalState.Inflation(0) = 0.07
    
bondedRatio = GlobalState.BondedPool / GlobalState.TotalSupply
AnnualInflationRateChange = (1 - bondedRatio / 0.67) * 0.13

annualInflation += AnnualInflationRateChange

if annualInflation > 0.20 then GlobalState.Inflation = 0.20
if annualInflation < 0.07 then GlobalState.Inflation = 0.07

provisionTokensHourly = GlobalState.TotalSupply * GlobalState.Inflation / (365.25*24)
```

Because the validators hold a relative bonded share (`GlobalStakeShares`), when
more bonded tokens are added proportionally to all validators, the only term
which needs to be updated is the `GlobalState.BondedPool`. So for each previsions
cycle:

```
GlobalState.BondedPool += provisionTokensHourly
```
