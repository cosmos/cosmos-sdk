## State

### Pool

 - key: `01`
 - value: `amino(pool)`

The pool is a space for all dynamic global state of the Cosmos Hub.  It tracks
information about the total amounts of Atoms in all states, representative
validator shares for stake in the global pools, moving Atom inflation
information, etc.

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
 - key: `00`
 - value: `amino(params)`

Params is global data structure that stores system parameters and defines
overall functioning of the stake module. 

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

Validators are identified according to the `ValOwnerAddr`, 
an SDK account address for the owner of the validator.

Validators also have a `ValTendermintAddr`, the address 
of the public key of the validator.

Validators are indexed in the store using the following maps:

 - Validators: `0x02 | ValOwnerAddr -> amino(validator)`
 - ValidatorsByPubKey: `0x03 | ValTendermintAddr -> ValOwnerAddr`
 - ValidatorsByPower: `0x05 | power | blockHeight | blockTx  -> ValOwnerAddr`

 `Validators` is the primary index - it ensures that each owner can have only one
 associated validator, where the public key of that validator can change in the
 future. Delegators can refer to the immutable owner of the validator, without
 concern for the changing public key.

 `ValidatorsByPubKey` is a secondary index that enables lookups for slashing.
 When Tendermint reports evidence, it provides the validator address, so this
 map is needed to find the owner.

 `ValidatorsByPower` is a secondary index that provides a sorted list of
 potential validators to quickly determine the current active set. For instance,
 the first 100 validators in this list can be returned with every EndBlock.

The `Validator` holds the current state and some historical actions of the
validator.

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

Delegations are identified by combining `DelegatorAddr` (the address of the delegator) with the ValOwnerAddr 
Delegators are indexed in the store as follows:

 - Delegation: ` 0x0A | DelegatorAddr | ValOwnerAddr -> amino(delegation)`

Atom holders may delegate coins to validators; under this circumstance their
funds are held in a `Delegation` data structure. It is owned by one 
delegator, and is associated with the shares for one validator. The sender of 
the transaction is the owner of the bond.

```golang
type Delegation struct {
	Shares        sdk.Rat      // delegation shares recieved 
	Height        int64        // last height bond updated
}
```

### UnbondingDelegation

Shares in a `Delegation` can be unbonded, but they must for some time exist as an `UnbondingDelegation`,
where shares can be reduced if Byzantine behaviour is detected.

`UnbondingDelegation` are indexed in the store as:

 - UnbondingDelegationByDelegator: ` 0x0B | DelegatorAddr | ValOwnerAddr ->
   amino(unbondingDelegation)`
 - UnbondingDelegationByValOwner: ` 0x0C | ValOwnerAddr | DelegatorAddr | ValOwnerAddr ->
   nil`

 The first map here is used in queries, to lookup all unbonding delegations for
 a given delegator, while the second map is used in slashing, to lookup all
 unbonding delegations associated with a given validator that need to be
 slashed.

A UnbondingDelegation object is created every time an unbonding is initiated.
The unbond must be completed with a second transaction provided by the
delegation owner after the unbonding period has passed.

```golang
type UnbondingDelegation struct {
    Tokens           sdk.Coins   // the value in Atoms of the amount of shares which are unbonding
    CompleteTime     int64       // unix time to complete redelegation
}
``` 

### Redelegation

Shares in a `Delegation` can be rebonded to a different validator, but they must for some time exist as a `Redelegation`,
where shares can be reduced if Byzantine behaviour is detected. This is tracked
as moving a delegation from a `FromValOwnerAddr` to a `ToValOwnerAddr`.

`Redelegation` are indexed in the store as:

 - Redelegations: `0x0D | DelegatorAddr | FromValOwnerAddr | ToValOwnerAddr ->
   amino(redelegation)`
 - RedelegationsBySrc: `0x0E | FromValOwnerAddr | ToValOwnerAddr |
   DelegatorAddr -> nil`
 - RedelegationsByDst: `0x0F | ToValOwnerAddr | FromValOwnerAddr | DelegatorAddr
   -> nil`


The first map here is used for queries, to lookup all redelegations for a given
delegator. The second map is used for slashing based on the FromValOwnerAddr,
while the third map is for slashing based on the ToValOwnerAddr.

A redelegation object is created every time a redelegation occurs. The
redelegation must be completed with a second transaction provided by the
delegation owner after the unbonding period has passed.  The destination
delegation of a redelegation may not itself undergo a new redelegation until
the original redelegation has been completed.

```golang
type Redelegation struct {
    SourceShares           sdk.Rat     // amount of source shares redelegating
    DestinationShares      sdk.Rat     // amount of destination shares created at redelegation
    CompleteTime           int64       // unix time to complete redelegation
}
```
