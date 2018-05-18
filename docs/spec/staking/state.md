
## State

The staking module persists the following information to the store:
* `GlobalState`, a struct describing the global pools, inflation, and
  fees
* `ValidatorValidators: <pubkey | shares> => <validator>`, a map of all validators (including current validators) in the store,
indexed by their public key and shares in the global pool.
* `DelegatorBonds: < delegator-address | validator-pubkey > => <delegator-bond>`. a map of all delegations by a delegator to a validator,
indexed by delegator address and validator pubkey.
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
    UnbondedPool             int64        // reserve of unbonding tokens held with validators
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

### Validator

The `Validator` holds the current state and some historical 
actions of validators. 

``` go
type ValidatorStatus byte

const (
    Bonded   ValidatorStatus = 0x01
    Unbonded ValidatorStatus = 0x02
    Revoked  ValidatorStatus = 0x03
)

type Validator struct {
    Status                 ValidatorStatus       
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

Validator parameters are described:
* Status: it can be Bonded (active validator), Unbonding (validator) 
  or Revoked
* ConsensusPubKey: validator public key that is used strictly for participating in 
  consensus
* GovernancePubKey: public key used by the validator for governance voting 
* Owner: Address that is allowed to unbond coins.
* GlobalStakeShares: Represents shares of `GlobalState.BondedPool` if 
  `Validator.Status` is `Bonded`; or shares of `GlobalState.Unbondingt Pool` 
  otherwise
* IssuedDelegatorShares: Sum of all shares a validator issued to delegators 
  (which includes the validator's self-bond); a delegator share represents 
  their stake in the Validator's `GlobalStakeShares`
* RedelegatingShares: The portion of `IssuedDelegatorShares` which are 
  currently re-delegating to a new validator
* VotingPower: Proportional to the amount of bonded tokens which the validator
  has if `Validator.Status` is `Bonded`; otherwise it is equal to `0`
* Commission:  The commission rate of fees charged to any delegators
* CommissionMax:  The maximum commission rate this validator can charge each 
  day from the date `GlobalState.DateLastCommissionReset` 
* CommissionChangeRate: The maximum daily increase of the validator commission
* CommissionChangeToday: Counter for the amount of change to commission rate 
  which has occurred today, reset on the first block of each day (UTC time)
* ProposerRewardPool: reward pool for extra fees collected when this validator
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

Atom holders may delegate coins to validators; under this circumstance their
funds are held in a `DelegatorBond` data structure. It is owned by one 
delegator, and is associated with the shares for one validator. The sender of 
the transaction is the owner of the bond.

``` go
type DelegatorBond struct {
    Validator            crypto.PubKey
    Shares               rational.Rat
    AdjustmentFeePool    coin.Coins  
    AdjustmentRewardPool coin.Coins  
} 
```

Description: 
* Validator: the public key of the validator: bonding too
* Shares: the number of delegator shares received from the validator
* AdjustmentFeePool: Adjustment factor used to passively calculate each bonds
  entitled fees from `GlobalState.FeePool`
* AdjustmentRewardPool: Adjustment factor used to passively calculate each
  bonds entitled fees from `Validator.ProposerRewardPool`

 
### QueueElem

The Unbonding and re-delegation process is implemented using the ordered queue 
data structure. All queue elements share a common structure:

```golang
type QueueElem struct {
    Validator   crypto.PubKey
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
    StartSlashRatio  rational.Rat  // validator slash ratio 
}
``` 

### QueueElemReDelegate

QueueElemReDelegate structure is used in the re-delegation queue. 

```golang
type QueueElemReDelegate struct {
    QueueElem
    Payout       Address       // account to pay out to
    Shares       rational.Rat  // amount of shares which are unbonding
    NewValidator crypto.PubKey // validator to bond to after unbond
}
```

