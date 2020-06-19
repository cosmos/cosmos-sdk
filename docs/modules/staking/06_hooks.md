<!--
order: 5
-->

# Hooks

Other modules may register operations to execute when a certain event has
occurred within staking.  These events can be registered to execute either
right `Before` or `After` the staking event (as per the hook name). The
following hooks can registered with staking: 

 - `AfterValidatorCreated(Context, ValAddress)`
   - called when a validator is created
 - `BeforeValidatorModified(Context, ValAddress)`
   - called when a validator's state is changed
 - `AfterValidatorRemoved(Context, ConsAddress, ValAddress)`
   - called when a validator is deleted
 - `AfterValidatorBonded(Context, ConsAddress, ValAddress)`
   - called when a validator is bonded
 - `AfterValidatorBeginUnbonding(Context, ConsAddress, ValAddress)`
   - called when a validator begins unbonding
 - `BeforeDelegationCreated(Context, AccAddress, ValAddress)`
   - called when a delegation is created
 - `BeforeDelegationSharesModified(Context, AccAddress, ValAddress)`
   - called when a delegation's shares are modified
 - `BeforeDelegationRemoved(Context, AccAddress, ValAddress)`
   - called when a delegation is removed
