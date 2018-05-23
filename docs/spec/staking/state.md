
## State

The staking module persists the following information to the store:
* `Params`, a struct describing the global pools, inflation, and fees
* `Pool`, a struct describing the global pools, inflation, and fees
* `ValidatorValidators: <pubkey | shares> => <validator>`, a map of all validators (including current validators) in the store,
indexed by their public key and shares in the global pool.
* `DelegatorBonds: < delegator-address | validator-pubkey > => <delegator-bond>`. a map of all delegations by a delegator to a validator,
indexed by delegator address and validator pubkey.
  public key

### Pool

The pool is a space for all dynamic global state of the Cosmos Hub.  It tracks
information about the total amounts of Atoms in all states, representative
validator shares for stake in the global pools, moving Atom inflation
information, etc.

```golang
type Pool struct {
    LooseUnbondedTokens int64    // tokens not associated with any validator
	UnbondedTokens      int64    // reserve of unbonded tokens held with validators
	UnbondingTokens     int64    // tokens moving from bonded to unbonded pool
	BondedTokens        int64    // reserve of bonded tokens
	UnbondedShares      sdk.Rat  // sum of all shares distributed for the Unbonded Pool
	UnbondingShares     sdk.Rat  // shares moving from Bonded to Unbonded Pool
	BondedShares        sdk.Rat  // sum of all shares distributed for the Bonded Pool
	InflationLastTime   int64    // block which the last inflation was processed // TODO make time
	Inflation           sdk.Rat  // current annual inflation rate

	DateLastCommissionReset int64  // unix timestamp for last commission accounting reset (daily)
}

type PoolShares struct {
	Status sdk.BondStatus  // either: unbonded, unbonding, or bonded
	Amount sdk.Rat         // total shares of type ShareKind
}
```

### Params

Params is global data structure that stores system parameters and defines
overall functioning of the stake module. 

```golang
type Params struct {
    InflationRateChange sdk.Rat  // maximum annual change in inflation rate
	InflationMax        sdk.Rat  // maximum inflation rate
	InflationMin        sdk.Rat  // minimum inflation rate
	GoalBonded          sdk.Rat  // Goal of percent bonded atoms

	MaxValidators uint16  // maximum number of validators
	BondDenom     string  // bondable coin denomination
}
```

### Validator

The `Validator` holds the current state and some historical actions of the
validator.

```golang
type Validator struct {
	Owner           sdk.Address    // sender of BondTx - UnbondTx returns here
	ConsensusPubKey crypto.PubKey  // Tendermint consensus pubkey of validator
	Revoked         bool           // has the validator been revoked?

	PoolShares      PoolShares  // total shares for tokens held in the pool
	DelegatorShares sdk.Rat     // total shares issued to a validator's delegators

	Description        Description  // description terms for the validator
	BondHeight         int64        // earliest height as a bonded validator
	BondIntraTxCounter int16        // block-local tx index of validator change
	ProposerRewardPool sdk.Coins    // reward pool collected from being the proposer

	Commission            sdk.Rat  // the commission rate of fees charged to any delegators
	CommissionMax         sdk.Rat  // maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Rat  // maximum daily increase of the validator commission
	CommissionChangeToday sdk.Rat  // commission rate change today, reset each day (UTC time)

	PrevPoolShares PoolShares  // total shares of a global hold pools
}

type Description struct {
	Moniker  string  // name
	Identity string  // optional identity signature (ex. UPort or Keybase)
	Website  string  // optional website link
	Details  string  // optional details
}
```

* RedelegatingShares: The portion of `IssuedDelegatorShares` which are 
  currently re-delegating to a new validator

### Delegation

Atom holders may delegate coins to validators; under this circumstance their
funds are held in a `Delegation` data structure. It is owned by one 
delegator, and is associated with the shares for one validator. The sender of 
the transaction is the owner of the bond.

```golang
type Delegation struct {
	DelegatorAddr sdk.Address  // delegation owner address
	ValidatorAddr sdk.Address  // validator owner address
	Shares        sdk.Rat      // delegation shares recieved 
	Height        int64        // last height bond updated
}
```

### UnbondingDelegation

A UnbondingDelegation object is created every time an unbonding is initiated.
It must be completed with a second transaction provided by the delegation owner
after the unbonding period has passed.

```golang
type UnbondingDelegation struct {
    DelegationKey    []byte     // key of the delegation
    InitTime         int64      // unix time at unbonding initation
    InitHeight       int64      // block height at unbonding initation
    ExpectedTokens   sdk.Coins  // the value in Atoms of the amount of shares which are unbonding
    StartSlashRatio  sdk.Rat    // validator slash ratio at unbonding initiation
}
``` 

### Redelegation

A redelegation object is created every time a redelegation occurs. It must be
completed with a second transaction provided by the delegation owner after the
unbonding period has passed.  The destination delegation of a redelegation may
not itself undergo a new redelegation until the original redelegation has been
completed.

 - index: delegation address
 - index 2: source validator owner address
 - index 3: destination validator owner address

```golang
type Redelegation struct {
    SourceDelegation       []byte  // source delegation key
    DestinationDelegation  []byte  // destination delegation key
    InitTime     int64             // unix time at redelegation
    InitHeight   int64             // block height at redelegation
    Shares       sdk.Rat           // amount of shares redelegating
}
```
