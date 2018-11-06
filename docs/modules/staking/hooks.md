## Receiver Hooks

The staking module allow for the following hooks to be registered with staking events:

``` golang
// event hooks for staking validator object
type StakingHooks interface {
	OnValidatorCreated(ctx Context, address ValAddress)  // Must be called when a validator is created
	OnValidatorModified(ctx Context, address ValAddress) // Must be called when a validator's state changes
	OnValidatorRemoved(ctx Context, address ValAddress)  // Must be called when a validator is deleted

	OnValidatorBonded(ctx Context, address ConsAddress)         // called when a validator is bonded
	OnValidatorBeginUnbonding(ctx Context, address ConsAddress, operator ValAddress) // called when a validator begins unbonding

	OnDelegationCreated(ctx Context, delAddr AccAddress, valAddr ValAddress)        // called when a delegation is created
	OnDelegationSharesModified(ctx Context, delAddr AccAddress, valAddr ValAddress) // called when a delegation's shares are modified
	OnDelegationRemoved(ctx Context, delAddr AccAddress, valAddr ValAddress)        // called when a delegation is removed
}
```
