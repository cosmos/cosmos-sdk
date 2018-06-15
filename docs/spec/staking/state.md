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

val owner address: SDK account addresss of the owner of the validator :)
tm val pubkey: Public Key of the Tendermint Validator 

 - map1: <val owner address> -> <validator>
 - map2: <tm val address> -> <val owner address>  
 - map3: <power | block height | block tx > -> <val owner address> 

 map1 is the main lookup. each owner can have only one validator.
 delegators point to an immutable owner
 owners can change their TM val pubkey
 need map2 so we can do lookups for slashing !
 need map3 so we have sorted vals to know the top 100

-----------

The `Validator` holds the current state and some historical actions of the
validator.

 - stored object:

```golang
type Validator struct {
	ConsensusPubKey crypto.PubKey  // Tendermint consensus pubkey of validator
	Revoked         bool           // has the validator been revoked?

	PoolShares      PoolShares     // total shares for tokens held in the pool
	DelegatorShares sdk.Rat        // total shares issued to a validator's delegators
	SlashRatio      sdk.Rat        // increases each time the validator is slashed

	Description        Description  // description terms for the validator

    // Needed for ordering vals in the bypower key
	BondHeight         int64        // earliest height as a bonded validator
	BondIntraTxCounter int16        // block-local tx index of validator change

    CommissionInfo      CommissionInfo // info about the validator's commission

	ProposerRewardPool sdk.Coins    // reward pool collected from being the proposer

    // TODO: maybe this belongs in distribution module ?
	PrevPoolShares PoolShares  // total shares of a global hold pools
}

type CommissionInfo struct {
	Rate        sdk.Rat  // the commission rate of fees charged to any delegators
	Max         sdk.Rat  // maximum commission rate which this validator can ever charge
	ChangeRate  sdk.Rat  // maximum daily increase of the validator commission
	ChangeToday sdk.Rat  // commission rate change today, reset each day (UTC time)
    LastChange  int64    // unix timestamp of last commission change
}

type Description struct {
	Moniker  string // name
	Identity string // optional identity signature (ex. UPort or Keybase)
	Website  string // optional website link
	Details  string // optional details
}
```

### Delegation

 - map1: < delegator address | val owner address > -> < delegation >

Atom holders may delegate coins to validators; under this circumstance their
funds are held in a `Delegation` data structure. It is owned by one 
delegator, and is associated with the shares for one validator. The sender of 
the transaction is the owner of the bond.

 - stored object:

```golang
type Delegation struct {
	Shares        sdk.Rat      // delegation shares recieved 
	Height        int64        // last height bond updated
}
```

### UnbondingDelegation

 - map1: < prefix-unbonding | delegator address | val owner address > -> < unbonding delegation >
 - map2: < prefix-unbonding | val owner address | delegator address > -> nil

 map1 for queries.
 map2 for eager slashing

A UnbondingDelegation object is created every time an unbonding is initiated.
The unbond must be completed with a second transaction provided by the
delegation owner after the unbonding period has passed.

 - stored object:

```golang
type UnbondingDelegation struct {
    Tokens           sdk.Coins   // the value in Atoms of the amount of shares which are unbonding
    CompleteTime     int64       // unix time to complete redelegation
}
``` 

### Redelegation

 - map1: < prefix-redelegation | delegator address | from val owner address | to
   val owner address > -> < redelegation >
 - map2: < prefix-redelegation | from val owner address | to
   val owner address | delegator > -> nil
 - map2: < prefix-redelegation | to val owner address | from
   val owner address | delegator > -> nil

 map1: queries
 map2: slash for from validator
 map3: slash for to validator

A redelegation object is created every time a redelegation occurs. The
redelegation must be completed with a second transaction provided by the
delegation owner after the unbonding period has passed.  The destination
delegation of a redelegation may not itself undergo a new redelegation until
the original redelegation has been completed.

 - stored object:

```golang
type Redelegation struct {
    SourceShares           sdk.Rat     // amount of source shares redelegating
    DestinationShares      sdk.Rat     // amount of destination shares created at redelegation
    CompleteTime           int64       // unix time to complete redelegation
}
```
