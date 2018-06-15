## State

### Pool
 - index: n/a single-record

The pool is a space for all dynamic global state of the Cosmos Hub.  It tracks
information about the total amounts of Atoms in all states, representative
validator shares for stake in the global pools, moving Atom inflation
information, etc.

 - stored object:

```golang
type Pool struct {
    LooseUnbondedTokens int64   // tokens not associated with any validator
	UnbondedTokens      int64   // reserve of unbonded tokens held with validators
	UnbondingTokens     int64   // tokens moving from bonded to unbonded pool
	BondedTokens        int64   // reserve of bonded tokens
	UnbondedShares      sdk.Rat // sum of all shares distributed for the Unbonded Pool
	UnbondingShares     sdk.Rat // shares moving from Bonded to Unbonded Pool
	BondedShares        sdk.Rat // sum of all shares distributed for the Bonded Pool
	InflationLastTime   int64   // block which the last inflation was processed // TODO make time
	Inflation           sdk.Rat // current annual inflation rate

	DateLastCommissionReset int64  // unix timestamp for last commission accounting reset (daily)
}

type PoolShares struct {
	Status sdk.BondStatus // either: unbonded, unbonding, or bonded
	Amount sdk.Rat        // total shares of type ShareKind
}
```

### Params
 - index: n/a single-record

Params is global data structure that stores system parameters and defines
overall functioning of the stake module. 

 - stored object:

```golang
type Params struct {
    InflationRateChange sdk.Rat // maximum annual change in inflation rate
	InflationMax        sdk.Rat // maximum inflation rate
	InflationMin        sdk.Rat // minimum inflation rate
	GoalBonded          sdk.Rat // Goal of percent bonded atoms

	MaxValidators uint16 // maximum number of validators
	BondDenom     string // bondable coin denomination
}
```

### Validator
 - index 1: validator owner address
 - index 2: validator Tendermint PubKey
 - index 3: bonded validators only
 - index 4: voting power

Related Store which holds Validator.ABCIValidator()
 - index: validator owner address

The `Validator` holds the current state and some historical actions of the
validator.

 - stored object:

```golang
type Validator struct {
	Owner           sdk.Address    // sender of BondTx - UnbondTx returns here
	ConsensusPubKey crypto.PubKey  // Tendermint consensus pubkey of validator
	Revoked         bool           // has the validator been revoked?

	PoolShares      PoolShares     // total shares for tokens held in the pool
	DelegatorShares sdk.Rat        // total shares issued to a validator's delegators
	SlashRatio      sdk.Rat        // increases each time the validator is slashed

	Description        Description // description terms for the validator
	BondHeight         int64       // earliest height as a bonded validator
	BondIntraTxCounter int16       // block-local tx index of validator change
	ProposerRewardPool sdk.Coins   // reward pool collected from being the proposer

	Commission            sdk.Rat  // the commission rate of fees charged to any delegators
	CommissionMax         sdk.Rat  // maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Rat  // maximum daily increase of the validator commission
	CommissionChangeToday sdk.Rat  // commission rate change today, reset each day (UTC time)

	PrevPoolShares PoolShares      // total shares of a global hold pools
}

type Description struct {
	Moniker  string // name
	Identity string // optional identity signature (ex. UPort or Keybase)
	Website  string // optional website link
	Details  string // optional details
}
```

### Delegation
 - index: delegation address

Atom holders may delegate coins to validators; under this circumstance their
funds are held in a `Delegation` data structure. It is owned by one 
delegator, and is associated with the shares for one validator. The sender of 
the transaction is the owner of the bond.

 - stored object:

```golang
type Delegation struct {
	DelegatorAddr sdk.Address // delegation owner address
	ValidatorAddr sdk.Address // validator owner address
	Shares        sdk.Rat     // delegation shares recieved 
	Height        int64       // last height bond updated
}
```

### UnbondingDelegation
 - index: delegation address

A UnbondingDelegation object is created every time an unbonding is initiated.
The unbond must be completed with a second transaction provided by the
delegation owner after the unbonding period has passed.

 - stored object:

```golang
type UnbondingDelegation struct {
    DelegationKey    sdk.Address // key of the delegation
    ExpectedTokens   sdk.Coins   // the value in Atoms of the amount of shares which are unbonding
    StartSlashRatio  sdk.Rat     // validator slash ratio at unbonding initiation
    CompleteTime     int64       // unix time to complete redelegation
    CompleteHeight   int64       // block height to complete redelegation
}
``` 

### Redelegation
 - index 1: delegation address
 - index 2: source validator owner address
 - index 3: destination validator owner address

A redelegation object is created every time a redelegation occurs. The
redelegation must be completed with a second transaction provided by the
delegation owner after the unbonding period has passed.  The destination
delegation of a redelegation may not itself undergo a new redelegation until
the original redelegation has been completed.

 - stored object:

```golang
type Redelegation struct {
    SourceDelegation       sdk.Address // source delegation key
    DestinationDelegation  sdk.Address // destination delegation key
    SourceShares           sdk.Rat     // amount of source shares redelegating
    DestinationShares      sdk.Rat     // amount of destination shares created at redelegation
    SourceStartSlashRatio  sdk.Rat     // source validator slash ratio at unbonding initiation
    CompleteTime           int64       // unix time to complete redelegation
    CompleteHeight         int64       // block height to complete redelegation
}
```
