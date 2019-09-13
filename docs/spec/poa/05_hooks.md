# Hooks

Other modules may register operations to execute when a certain event has
occurred within the POA module. These events can be registered to execute either
right `Before` or `After` the POA event (as per the hook name). The
following hooks can registered with the POA module:

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
