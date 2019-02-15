# State Transitions

This document describes the state transition operations pertaining to:

 - Validators
 - Delegations
 - Slashing


## Validators

### Non-Bonded to Bonded

When a validator is bonded from any other state the following operations occur:  
 - set `validator.Status` to `Bonded`
 - update the `Pool` object with tokens moved from `NotBondedTokens` to `BondedTokens`
 - delete record the existing record from `ValidatorByPowerIndex`
 - add an new updated record to the `ValidatorByPowerIndex`
 - update the `Validator` object for this validator
 - if it exists, delete any `ValidatorQueue` record for this validator 

### Bonded to Unbonding
When a validator begins the unbonding process the following operations occur: 
 - update the `Pool` object with tokens moved from `BondedTokens` to `NotBondedTokens`
 - set `validator.Status` to `Unbonding`
 - delete record the existing record from `ValidatorByPowerIndex`
 - add an new updated record to the `ValidatorByPowerIndex`
 - update the `Validator` object for this validator
 - insert a new record into the `ValidatorQueue` for this validator 

### Unbonding to Unbonded
A validator moves from unbonding to unbonded when the `ValidatorQueue` object
moves from bonded to unbonded
 - update the `Validator` object for this validator
 - set `validator.Status` to `Unbonded`

### Jail/Unjail 
when a validator is jailed it is effectively removed from the Tendermint set.
this process may be also be reversed. the following operations occur:
 - set `Validator.Jailed` and update object 
 - if jailed delete record from `ValidatorByPowerIndex`
 - if unjailed add record to `ValidatorByPowerIndex`


## Delegations

### Delegate
When a delegation occurs both the validator and the delegtion objects are affected  
 - determine the delegators shares based on tokens delegated and the validator's exchange rate
 - remove tokens from the sending account 
 - add shares the delegation object or add them to a created validator object
 - add new delegator shares and update the `Validator` object
 - update the `Pool` object appropriately if tokens have moved into a bonded validator
 - delete record the existing record from `ValidatorByPowerIndex`
 - add an new updated record to the `ValidatorByPowerIndex`

#### Unbond Delegation
As a part of the Undelegate and Complete Unbonding state transitions Unbond
Delegation may be called. 
 - subtract the unbonded shares from delegator
 - update the delegation or remove the delegation if there are no more shares
 - if the delegation is the operator of the validator and no more shares exist
   then trigger a jail validator
 - update the validator with removed the delegator shares and associated coins, update
   the pool for any shifts between bonded and non-bonded tokens. 
 - remove the validator if it is unbonded and there are no more delegation shares. 

### Undelegate
When an delegation occurs both the validator and the delegtion objects are affected  
 - perform an unbond delegation 
   - if the validator is unbonding or bonded add the tokens to an
     `UnbondingDelegation` Entry
   - if the validator is unbonded send the tokens directly to the withdraw
     account

### Complete Unbonding
For undelegations which do not complete immediately, the following operations
occur when the unbonding delegation queue element matures:
 - remove the entry from the `UnbondingDelegation` object
 - withdraw the tokens to the delegator withdraw address 

### Begin Redelegation
Redelegations affect the delegation, source and destination validators. 
 - perform an unbond delegation from the source validator 
 - using the generated tokens perform a Delegate to the destination
   validator
 - record the token amount in an new entry in the relevant `Redelegation`

### Complete Redelegation
When a redelegations complete the following occurs:
 - remove the entry from the `Redelegation` object


TODO TODO TOFU TODO
## Slashing

### Slash Validator

### Slash Unbonding Delegation

### Slash Redelegation
