## State

### Pool

The pool is a space for all dynamic global state of the Cosmos Hub.  It tracks
information about the total amounts of Atoms in all states, moving Atom
inflation information, etc.

 - Pool: `0x01 -> amino(pool)`

```golang
type Pool struct {
    LooseTokens         int64   // tokens not associated with any bonded validator
    BondedTokens        int64   // reserve of bonded tokens
}
```

### Params

Params is global data structure that stores system parameters and defines
overall functioning of the stake module. 

 - Params: `0x00 -> amino(params)`

```golang
type Params struct {
    MaxValidators uint16 // maximum number of validators
    BondDenom     string // bondable coin denomination
}
```

### Validator

Validators are identified according to the `OperatorAddr`, an SDK validator
address for the operator of the validator.

Validators also have a `ConsPubKey`, the public key of the validator used in
Tendermint consensus. The validator can be retrieved from it's `ConsPubKey`
once it can be converted into the corresponding `ConsAddr`. Validators are
indexed in the store using the following maps:

- Validators: `0x02 | OperatorAddr -> amino(validator)`
- ValidatorsByConsAddr: `0x03 | ConsAddr -> OperatorAddr`
- ValidatorsByPower: `0x05 | power | blockHeight | blockTx  -> OperatorAddr`

`Validators` is the primary index - it ensures that each operator can have only one
associated validator, where the public key of that validator can change in the
future. Delegators can refer to the immutable operator of the validator, without
concern for the changing public key.

`ValidatorsByPubKey` is a secondary index that enables lookups for slashing.
When Tendermint reports evidence, it provides the validator address, so this
map is needed to find the operator.

`ValidatorsByPower` is a secondary index that provides a sorted list of
potential validators to quickly determine the current active set. For instance,
the first 100 validators in this list can be returned with every EndBlock.

The `Validator` holds the current state and some historical actions of the
validator.

```golang
type Validator struct {
    ConsPubKey      crypto.PubKey  // Tendermint consensus pubkey of validator
    Jailed          bool           // has the validator been jailed?

    Status          sdk.BondStatus // validator status (bonded/unbonding/unbonded)
    Tokens          sdk.Dec        // delegated tokens (incl. self-delegation)
    DelegatorShares sdk.Dec        // total shares issued to a validator's delegators
    SlashRatio      sdk.Dec        // increases each time the validator is slashed

    Description        Description  // description terms for the validator

    // Needed for ordering vals in the by-power key
    BondHeight         int64        // earliest height as a bonded validator
    BondIntraTxCounter int16        // block-local tx index of validator change

    CommissionInfo     CommissionInfo // info about the validator's commission
}

type CommissionInfo struct {
    Rate        sdk.Dec  // the commission rate of fees charged to any delegators
    Max         sdk.Dec  // maximum commission rate which this validator can ever charge
    ChangeRate  sdk.Dec  // maximum daily increase of the validator commission
    ChangeToday sdk.Dec  // commission rate change today, reset each day (UTC time)
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

Delegations are identified by combining `DelegatorAddr` (the address of the delegator)
with the `OperatorAddr` Delegators are indexed in the store as follows:

- Delegation: ` 0x0A | DelegatorAddr | OperatorAddr -> amino(delegation)`

Atom holders may delegate coins to validators; under this circumstance their
funds are held in a `Delegation` data structure. It is owned by one
delegator, and is associated with the shares for one validator. The sender of
the transaction is the owner of the bond.

```golang
type Delegation struct {
    Shares        sdk.Dec   // delegation shares received
    Height        int64     // last height bond updated
}
```

### UnbondingDelegation

Shares in a `Delegation` can be unbonded, but they must for some time exist as an `UnbondingDelegation`, where shares can be reduced if Byzantine behavior is detected.

`UnbondingDelegation` are indexed in the store as:

- UnbondingDelegationByDelegator: ` 0x0B | DelegatorAddr | OperatorAddr ->
   amino(unbondingDelegation)`
- UnbondingDelegationByValOwner: ` 0x0C | OperatorAddr | DelegatorAddr | OperatorAddr ->
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

Shares in a `Delegation` can be rebonded to a different validator, but they must
for some time exist as a `Redelegation`, where shares can be reduced if Byzantine
behavior is detected. This is tracked as moving a delegation from a `FromOperatorAddr`
to a `ToOperatorAddr`.

`Redelegation` are indexed in the store as:

 - Redelegations: `0x0D | DelegatorAddr | FromOperatorAddr | ToOperatorAddr ->
   amino(redelegation)`
 - RedelegationsBySrc: `0x0E | FromOperatorAddr | ToOperatorAddr |
   DelegatorAddr -> nil`
 - RedelegationsByDst: `0x0F | ToOperatorAddr | FromOperatorAddr | DelegatorAddr
   -> nil`

The first map here is used for queries, to lookup all redelegations for a given
delegator. The second map is used for slashing based on the `FromOperatorAddr`,
while the third map is for slashing based on the ToValOwnerAddr.

A redelegation object is created every time a redelegation occurs. The
redelegation must be completed with a second transaction provided by the
delegation owner after the unbonding period has passed.  The destination
delegation of a redelegation may not itself undergo a new redelegation until
the original redelegation has been completed.

```golang
type Redelegation struct {
    SourceShares           sdk.Dec     // amount of source shares redelegating
    DestinationShares      sdk.Dec     // amount of destination shares created at redelegation
    CompleteTime           int64       // unix time to complete redelegation
}
```
