## Receiver Hooks

The staking module allow for the following hooks to be registered with staking events:

``` golang
type StakingHooks interface {
	AfterValidatorCreated(ctx Context, valAddr ValAddress)                       // Must be called when a validator is created
	BeforeValidatorModified(ctx Context, valAddr ValAddress)                      // Must be called when a validator's state changes
	AfterValidatorRemoved(ctx Context, consAddr ConsAddress, valAddr ValAddress) // Must be called when a validator is deleted

	AfterValidatorBonded(ctx Context, consAddr ConsAddress, valAddr ValAddress)         // Must be called when a validator is bonded
	AfterValidatorBeginUnbonding(ctx Context, consAddr ConsAddress, valAddr ValAddress) // Must be called when a validator begins unbonding
	AfterValidatorPowerDidChange(ctx Context, consAddr ConsAddress, valAddr ValAddress) // Called at EndBlock when a validator's power did change

	BeforeDelegationCreated(ctx Context, delAddr AccAddress, valAddr ValAddress)        // Must be called when a delegation is created
	BeforeDelegationSharesModified(ctx Context, delAddr AccAddress, valAddr ValAddress) // Must be called when a delegation's shares are modified
	BeforeDelegationRemoved(ctx Context, delAddr AccAddress, valAddr ValAddress)        // Must be called when a delegation is removed
}
```
