# Stake Module

## Overview

The stake module is tasked with various core staking functionality,
including validator set rotation, unbonding periods, and the 
distribution of inflationary provisions and transaction fees.
It is designed to efficiently facilitate small numbers of
validators (hundreds), and large numbers of delegators (tens of thousands).

Bonded Atoms are pooled globally and for each validator.
Validators have shares in the global pool, and delegators
have shares in the pool of every validator they delegate to.
Atom provisions simply accumulate in the global pool, making 
each share worth proportionally more.

Validator shares can be redeemed for Atoms, but the Atoms will be locked in a queue
for an unbonding period before they can be withdrawn to an account.
Delegators can exchange one validator's shares for another immediately
(ie. they can re-delegate to another validator), but must then wait the 
unbonding period before they can do it again.

Fees are pooled separately and withdrawn lazily, at any time. 
They are not bonded, and can be paid in multiple tokens.
An adjustment factor is maintained for each validator 
and delegator to determine the true proportion of fees in the pool they are entitled too. 
Adjustment factors are updated every time a validator or delegator's voting power changes.
Validators and delegators must withdraw all fees they are entitled too before they can bond or 
unbond Atoms.

## State

The staking module persists the following to the store:
- `GlobalState`, describing the global pools
- a `Candidate` for each candidate validator, indexed by public key
- a `Candidate` for each candidate validator, indexed by shares in the global pool (ie. ordered)
- a `DelegatorBond` for each delegation to a candidate by a delegator, indexed by delegator and candidate
    public keys
- a `Queue` of unbonding delegations (TODO)

### Global State

``` golang
type GlobalState struct {
	TotalSupply              int64        // total supply of atom tokens
	BondedShares             rational.Rat // sum of all shares distributed for the BondedPool
	UnbondedShares           rational.Rat // sum of all shares distributed for the UnbondedPool
	BondedPool               int64        // reserve of bonded tokens
	UnbondedPool             int64        // reserve of unbonded tokens held with candidates
	InflationLastTime        int64        // timestamp of last processing of inflation
	Inflation                rational.Rat // current annual inflation rate
    DateLastCommissionReset  int64        // unix timestamp for last commission accounting reset
    FeePool                  coin.Coins   // fee pool for all the fee shares which have already been distributed
    ReservePool              coin.Coins   // pool of reserve taxes collected on all fees for governance use
    Adjustment               rational.Rat // Adjustment factor for calculating global fee accum
}
```

### Candidate

The `Candidate` struct holds the current state and some historical actions of
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

type CandidateStatus byte
const (
    VyingUnbonded  CandidateStatus = 0x00
    VyingUnbonding CandidateStatus = 0x01
    Bonded         CandidateStatus = 0x02
    KickUnbonding  CandidateStatus = 0x03
    KickUnbonded   CandidateStatus = 0x04
)

type Description struct {
	Name       string 
	DateBonded string 
	Identity   string 
	Website    string 
	Details    string 
}
```

Candidate parameters are described:
 - Status: signal that the candidate is either vying for validator status
   either unbonded or unbonding, an active validator, or a kicked validator
   either unbonding or unbonded.
 - PubKey: separated key from the owner of the candidate as is used strictly
   for participating in consensus.
 - Owner: Address where coins are bonded from and unbonded to 
 - GlobalStakeShares: Represents shares of `GlobalState.BondedPool` if
   `Candidate.Status` is `Bonded`; or shares of `GlobalState.UnbondedPool` if
   `Candidate.Status` is otherwise
 - IssuedDelegatorShares: Sum of all shares issued to delegators (which
   includes the candidate's self-bond) which represent each of their stake in
   the Candidate's `GlobalStakeShares`
 - RedelegatingShares: The portion of `IssuedDelegatorShares` which are
   currently re-delegating to a new validator
 - VotingPower: Proportional to the amount of bonded tokens which the validator
   has if the validator is within the top 100 validators.
 - Commission:  The commission rate of fees charged to any delegators
 - CommissionMax:  The maximum commission rate which this candidate can charge 
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


Candidates are indexed by their `Candidate.PubKey`.
Additionally, we index empty values by the candidates global stake shares concatenated with the public key.

TODO: be more precise.

When the set of all validators needs to be determined from the group of all
candidates, the top candidates, sorted by GlobalStakeShares can be retrieved
from this sorting without the need to retrieve the entire group of candidates.
When validators are kicked from the validator set they are removed from this
list. 


### DelegatorBond

Atom holders may delegate coins to validators, under this circumstance their
funds are held in a `DelegatorBond`. It is owned by one delegator, and is
associated with the shares for one validator. The sender of the transaction is
considered to be the owner of the bond,  

``` golang
type DelegatorBond struct {
	Candidate            crypto.PubKey
	Shares               rational.Rat
    AdjustmentFeePool    coin.Coins  
    AdjustmentRewardPool coin.Coins  
} 
```

Description: 
 - Candidate: pubkey of the validator candidate: bonding too
 - Shares: the number of shares received from the validator candidate
 - AdjustmentFeePool: Adjustment factor used to passively calculate each bonds
   entitled fees from `GlobalState.FeePool`
 - AdjustmentRewardPool: Adjustment factor used to passively calculate each
   bonds entitled fees from `Candidate.ProposerRewardPool``

Each `DelegatorBond` is individually indexed within the store by delegator
address and candidate pubkey.

 - key: Delegator and Candidate-Pubkey
 - value: DelegatorBond 


### Unbonding Queue


- main unbonding queue contains both UnbondElem and RedelegateElem
    - "queue" + <i>
- new unbonding queue every time a val leaves the validator set
    - "queue"+ <candidate.pubkey > + <i>







The queue is ordered so the next to unbond/re-delegate is at the head. Every
tick the head of the queue is checked and if the unbonding period has passed
since `InitHeight` commence with final settlement of the unbonding and pop the
queue. All queue elements used for unbonding share a common struct:

``` golang
type QueueElem struct {
	Candidate   crypto.PubKey
	InitHeight  int64    // when the queue was initiated
}
```

``` golang
type QueueElemUnbondCandidate struct {
	QueueElem
}
```



``` golang
type QueueElemUnbondDelegation struct {
	QueueElem
	Payout           Address  // account to pay out to
    Shares           rational.Rat  // amount of shares which are unbonding
    StartSlashRatio  rational.Rat  // candidate slash ratio at start of re-delegation
}
```



``` golang
type QueueElemReDelegate struct {
	QueueElem
	Payout       Address  // account to pay out to
    Shares       rational.Rat  // amount of shares which are unbonding
    NewCandidate crypto.PubKey // validator to bond to after unbond
}
```


Each `QueueElem` is persisted in the store until it is popped from the queue. 

## Transactions

### TxDeclareCandidacy

Validator candidacy can be declared using the `TxDeclareCandidacy` transaction.
During this transaction a self-delegation transaction is executed to bond
tokens which are sent in with the transaction.

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


### TxLivelinessCheck

Liveliness kicks are only checked when a `TxLivelinessCheck` transaction is
submitted. 

``` golang
type TxLivelinessCheck struct {
    PubKey        crypto.PubKey
    RewardAccount Addresss
}
```

If the `TxLivelinessCheck is successful in kicking a validator, 5% of the
liveliness punishment is provided as a reward to `RewardAccount`.


### TxProveLive 

If the validator was kicked for liveliness issues and is able to regain
liveliness then all delegators in the temporary unbonding pool which have not
transacted to move will be bonded back to the now-live validator and begin to
once again collect provisions and rewards. Regaining livliness is demonstrated
by sending in a `TxProveLive` transaction:

``` golang
type TxProveLive struct {
    PubKey crypto.PubKey
}
```


### TxDelegate

All bonding, whether self-bonding or delegation, is done via
`TxDelegate`. 

Delegator bonds are created using the TxDelegate transaction. Within this
transaction the validator candidate queried with an amount of coins, whereby
given the current exchange rate of candidate's delegator-shares-to-atoms the
candidate will return shares which are assigned in `DelegatorBond.Shares`.

``` golang 
type TxDelegate struct { 
	PubKey crypto.PubKey
	Amount coin.Coin       
}
```

### TxUnbond 


In this context `TxUnbond` is used to
unbond either delegation bonds or validator self-bonds. 

Delegator unbonding is defined by the following transaction type:

``` golang
type TxUnbond struct { 
	PubKey crypto.PubKey
	Shares rational.Rat 
}
```


### TxRedelegate

The re-delegation command allows delegators to switch validators while still
receiving equal reward to as if you had never unbonded.

``` golang
type TxRedelegate struct { 
	PubKeyFrom crypto.PubKey
	PubKeyTo   crypto.PubKey
	Shares     rational.Rat 

}
```

A delegator who is in the process of unbonding from a validator may use the
re-delegate transaction to bond back to the original validator they're
currently unbonding from (and only that validator). If initiated, the delegator
will immediately begin to one again collect rewards from their validator. 

### TxWithdraw

....


## EndBlock

### Update Validators

The validator set is updated in the first block of every hour. Validators are
taken as the first `GlobalState.MaxValidators` number of candidates with the
greatest amount of staked atoms who have not been kicked from the validator
set.

Unbonding of an entire validator-candidate to a temporary liquid account occurs
under the scenarios: 
 - not enough stake to be within the validator set
 - the owner unbonds all of their staked tokens
 - validator liveliness issues
 - crosses a self-imposed safety threshold
   - minimum number of tokens staked by owner
   - minimum ratio of tokens staked by owner to delegator tokens 

When this occurs delegator's tokens do not unbond to their personal wallets but
begin the unbonding process to a pool where they must then transact in order to
withdraw to their respective wallets. 

### Unbonding

When unbonding is initiated, delegator shares are immediately removed from the
candidate and added to a queue object.

In the unbonding queue - the fraction of all historical slashings on
that validator are recorded (`StartSlashRatio`). When this queue reaches maturity
if that total slashing applied is greater on the validator then the
difference (amount that should have been slashed from the first validator) is
assigned to the amount being paid out. 


#### Liveliness issues

Liveliness issues are calculated by keeping track of the block precommits in
the block header. A queue is persisted which contains the block headers from
all recent blocks for the duration of the unbonding period. 

A validator is defined as having livliness issues if they have not been included in more than
33% of the blocks over: 
 - The most recent 24 Hours if they have >= 20% of global stake
 - The most recent week if they have = 0% of global stake
 - Linear interpolation of the above two scenarios


## Invariants

-----------------------------

------------




If a delegator chooses to initiate an unbond or re-delegation of their shares
while a candidate-unbond is commencing, then that unbond/re-delegation is
subject to a reduced unbonding period based on how much time those funds have
already spent in the unbonding queue.

### Re-Delegation

When re-delegation is initiated, delegator shares remain accounted for within
the `Candidate.Shares`, the term `RedelegatingShares` is incremented and a
queue element is created.

During the unbonding period all unbonding shares do not count towards the
voting power of a validator. Once the `QueueElemReDelegation` has reached
maturity, the appropriate unbonding shares are removed from the `Shares` and
`RedelegatingShares` term.   

Note that with the current menchanism a delegator cannot redelegate funds which
are currently redelegating. 

----------------------------------------------

## Provision Calculations

Every hour atom provisions are assigned proportionally to the each slashable
bonded token which includes re-delegating atoms but not unbonding tokens.

Validation provisions are payed directly to a global hold account
(`BondedTokenPool`) and proportions of that hold account owned by each
validator is defined as the `GlobalStakeBonded`. The tokens are payed as bonded
tokens.

Here, the bonded tokens that a candidate has can be calculated as:

```
globalStakeExRate = params.BondedTokenPool / params.IssuedGlobalStakeShares
candidateCoins = candidate.GlobalStakeShares * globalStakeExRate 
```

If a delegator chooses to add more tokens to a validator then the amount of
validator shares distributed is calculated on exchange rate (aka every
delegators shares do not change value at that moment. The validator's
accounting of distributed shares to delegators must also increased at every
deposit.
 
```
delegatorExRate = validatorCoins / candidate.IssuedDelegatorShares 
createShares = coinsDeposited / delegatorExRate 
candidate.IssuedDelegatorShares += createShares
```

Whenever a validator has new tokens added to it, the `BondedTokenPool` is
increased and must be reflected in the global parameter as well as the
validators `GlobalStakeShares`.  This calculation ensures that the worth of the
`GlobalStakeShares` of other validators remains worth a constant absolute
amount of the `BondedTokenPool`

```
createdGlobalStakeShares = coinsDeposited / globalStakeExRate 
validator.GlobalStakeShares +=  createdGlobalStakeShares
params.IssuedGlobalStakeShares +=  createdGlobalStakeShares

params.BondedTokenPool += coinsDeposited
```

Similarly, if a delegator wanted to unbond coins:

```
coinsWithdrawn = withdrawlShares * delegatorExRate

destroyedGlobalStakeShares = coinsWithdrawn / globalStakeExRate 
validator.GlobalStakeShares -= destroyedGlobalStakeShares
params.IssuedGlobalStakeShares -= destroyedGlobalStakeShares
params.BondedTokenPool -= coinsWithdrawn
```

Note that when an re-delegation occurs the shares to move are placed in an
re-delegation queue where they continue to collect validator provisions until
queue element matures. Although provisions are collected during re-delegation,
re-delegation tokens do not contribute to the voting power of a validator. 

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
annualInflation(0) = 0.07

bondedRatio = bondedTokenPool / totalTokenSupply
AnnualInflationRateChange = (1 - bondedRatio / 0.67) * 0.13

annualInflation += AnnualInflationRateChange

if annualInflation > 0.20 then annualInflation = 0.20
if annualInflation < 0.07 then annualInflation = 0.07

provisionTokensHourly = totalTokenSupply * annualInflation / (365.25*24)
```

Because the validators hold a relative bonded share (`GlobalStakeShare`), when
more bonded tokens are added proportionally to all validators the only term
which needs to be updated is the `BondedTokenPool`. So for each previsions
cycle:

```
params.BondedTokenPool += provisionTokensHourly
```

## Fee Calculations

Collected fees are pooled globally and divided out passively to validators and
delegators. Each validator has the opportunity to charge commission to the
delegators on the fees collected on behalf of the delegators by the validators.
Fees are paid directly into a global fee pool. Due to the nature of of passive
accounting whenever changes to parameters which affect the rate of fee
distribution occurs, withdrawal of fees must also occur. 
 
 - when withdrawing one must withdrawal the maximum amount they are entitled
   too, leaving nothing in the pool, 
 - when bonding, unbonding, or re-delegating tokens to an existing account a
   full withdrawal of the fees must occur (as the rules for lazy accounting
   change), 
 - when a candidate chooses to change the commission on fees, all accumulated 
   commission fees must be simultaneously withdrawn.

When the validator is the proposer of the round, that validator (and their
delegators) receives between 1% and 5% of fee rewards, the reserve tax is then
charged, then the remainder is distributed socially by voting power to all
validators including the proposer validator.  The amount of proposer reward is
calculated from pre-commits Tendermint messages. All provision rewards are
added to a provision reward pool which validator holds individually. Here note
that `BondedShares` represents the sum of all voting power saved in the
`GlobalState` (denoted `gs`).

```
proposerReward = feesCollected * (0.01 + 0.04 
                  * sumOfVotingPowerOfPrecommitValidators / gs.BondedShares)
candidate.ProposerRewardPool += proposerReward

reserveTaxed = feesCollected * params.ReserveTax
gs.ReservePool += reserveTaxed

distributedReward = feesCollected - proposerReward - reserveTaxed
gs.FeePool += distributedReward
gs.SumFeesReceived += distributedReward
gs.RecentFee = distributedReward
```

The entitlement to the fee pool held by the each validator can be accounted for
lazily.  First we must account for a candidate's `count` and `adjustment`. The
`count` represents a lazy accounting of what that candidates entitlement to the
fee pool would be if there `VotingPower` was to never change and they were to
never withdraw fees. 

``` 
candidate.count = candidate.VotingPower * BlockHeight
``` 

Similarly the GlobalState count can be passively calculated whenever needed,
where `BondedShares` is the updated sum of voting powers from all validators.

``` 
gs.count = gs.BondedShares * BlockHeight
``` 

The `adjustment` term accounts for changes in voting power and withdrawals of
fees. The adjustment factor must be persisted with the candidate and modified
whenever fees are withdrawn from the candidate or the voting power of the
candidate changes. When the voting power of the candidate changes the
`Adjustment` factor is increased/decreased by the cumulative difference in the
voting power if the voting power has been the new voting power as opposed to
the old voting power for the entire duration of the blockchain up the previous
block. Each time there is an adjustment change the GlobalState (denoted `gs`)
`Adjustment` must also be updated.

```
simplePool = candidate.count / gs.count * gs.SumFeesReceived
projectedPool = candidate.PrevPower * (height-1) 
                / (gs.PrevPower * (height-1)) * gs.PrevFeesReceived
                + candidate.Power / gs.Power * gs.RecentFee

AdjustmentChange = simplePool - projectedPool
candidate.AdjustmentRewardPool += AdjustmentChange
gs.Adjustment += AdjustmentChange
```

Every instance that the voting power changes, information about the state of
the validator set during the change must be recorded as a `powerChange` for
other validators to run through. Before any validator modifies its voting power
it must first run through the above calculation to determine the change in
their `caandidate.AdjustmentRewardPool` for all historical changes in the set
of `powerChange` which they have not yet synced to.  The set of all
`powerChange` may be trimmed from its oldest members once all validators have
synced past the height of the oldest `powerChange`.  This trim procedure will
occur on an epoch basis.  

```golang
type powerChange struct {
    height      int64        // block height at change
    power       rational.Rat // total power at change
    prevpower   rational.Rat // total power at previous height-1 
    feesin      coins.Coin   // fees in at block height
    prevFeePool coins.Coin   // total fees in at previous block height
}
```

Note that the adjustment factor may result as negative if the voting power of a
different candidate has decreased.  

``` 
candidate.AdjustmentRewardPool += withdrawn
gs.Adjustment += withdrawn
``` 

Now the entitled fee pool of each candidate can be lazily accounted for at 
any given block:

```
candidate.feePool = candidate.simplePool - candidate.Adjustment
```

So far we have covered two sources fees which can be withdrawn from: Fees from
proposer rewards (`candidate.ProposerRewardPool`), and fees from the fee pool
(`candidate.feePool`). However we should note that all fees from fee pool are
subject to commission rate from the owner of the candidate. These next
calculations outline the math behind withdrawing fee rewards as either a
delegator to a candidate providing commission, or as the owner of a candidate
who is receiving commission.

### Calculations For Delegators and Candidates

The same mechanism described to calculate the fees which an entire validator is
entitled to is be applied to delegator level to determine the entitled fees for
each delegator and the candidates entitled commission from `gs.FeesPool` and
`candidate.ProposerRewardPool`. 

The calculations are identical with a few modifications to the parameters:
 - Delegator's entitlement to `gs.FeePool`:
   - entitled party voting power should be taken as the effective voting power
     after commission is retrieved, 
     `bond.Shares/candidate.TotalDelegatorShares * candidate.VotingPower * (1 - candidate.Commission)`
 - Delegator's entitlement to `candidate.ProposerFeePool` 
   - global power in this context is actually shares
     `candidate.TotalDelegatorShares`
   - entitled party voting power should be taken as the effective shares after
     commission is retrieved, `bond.Shares * (1 - candidate.Commission)`
 - Candidate's commission entitlement to `gs.FeePool` 
   - entitled party voting power should be taken as the effective voting power
     of commission portion of total voting power, 
     `candidate.VotingPower * candidate.Commission`
 - Candidate's commission entitlement to `candidate.ProposerFeePool` 
   - global power in this context is actually shares
     `candidate.TotalDelegatorShares`
   - entitled party voting power should be taken as the of commission portion
     of total delegators shares, 
     `candidate.TotalDelegatorShares * candidate.Commission`

For more implementation ideas see spreadsheet `spec/AbsoluteFeeDistrModel.xlsx`

As mentioned earlier, every time the voting power of a delegator bond is
changing either by unbonding or further bonding, all fees must be
simultaneously withdrawn. Similarly if the validator changes the commission
rate, all commission on fees must be simultaneously withdrawn.  

### Other general notes on fees accounting

- When a delegator chooses to re-delegate shares, fees continue to accumulate
  until the re-delegation queue reaches maturity. At the block which the queue
  reaches maturity and shares are re-delegated all available fees are
  simultaneously withdrawn. 
- Whenever a totally new validator is added to the validator set, the `accum`
  of the entire candidate must be 0, meaning that the initial value for
  `candidate.Adjustment` must be set to the value of `canidate.Count` for the
  height which the candidate is added on the validator set.
- The feePool of a new delegator bond will be 0 for the height at which the bond
  was added. This is achieved by setting `DelegatorBond.FeeWithdrawalHeight` to
  the height which the bond was added. 
