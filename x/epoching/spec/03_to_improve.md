<!--
order: 3
-->

# Changes to make

## Validator self-unbonding (which exceed minimum self delegation) could be required to start instantly

Cases that trigger unbonding process

* Validator undelegate can unbond more tokens than his minimum_self_delegation and it will automatically turn the validator into unbonding
In this case, unbonding should start instantly.
* Validator miss blocks and get slashed
* Validator get slashed for double sign

**Note:** When a validator begins the unbonding process, it could be required to turn the validator into unbonding state instantly.
  This is different than a specific delegator beginning to unbond. A validator beginning to unbond means that it's not in the set any more.
  A delegator unbonding from a validator removes their delegation from the validator.

## Pending development

```go
// Changes to make
// — Implement correct next epoch time calculation
// — For validator self undelegation, it could be required to do start on end blocker
// — Implement TODOs on the PR #46
// Implement CLI commands for querying
// — BufferedValidators
// — BufferedMsgCreateValidatorQueue, BufferedMsgEditValidatorQueue
// — BufferedMsgUnjailQueue, BufferedMsgDelegateQueue, BufferedMsgRedelegationQueue, BufferedMsgUndelegateQueue
// Write epoch related tests with new scenarios
// — Simulation test is important for finding bugs [Ask Dev for questions)
// — Can easily add a simulator check to make sure all delegation amounts in queue add up to the same amount that’s in the EpochUnbondedPool
// — I’d like it added as an invariant test for the simulator
// — the simulator should check that the sum of all the queued delegations always equals the amount kept track in the data
// — Staking/Slashing/Distribution module params are being modified by governance based on vote result instantly. We should test the effect.
// — — Should test to see what would happen if max_validators is changed though, in the middle of an epoch
// — we should define some new invariants that help check that everything is working smoothly with these new changes for 3 modules e.g. https://github.com/cosmos/cosmos-sdk/blob/main/x/staking/keeper/invariants.go
// — — Within Epoch, ValidationPower = ValidationPower - SlashAmount
// — — When epoch actions queue is empty, EpochDelegationPool balance should be zero
// — we should count all the delegation changes that happen during the epoch, and then make sure that the resulting change at the end of the epoch is actually correct
// — If the validator that I delegated to double signs at block 16, I should still get slashed instantly because even though I asked to unbond at 14, they still used my power at block 16, I should only be not liable for slashes once my power is stopped being used
// — On the converse of this, I should still be getting rewards while my power is being used.  I shouldn’t stop receiving rewards until block 20
```
