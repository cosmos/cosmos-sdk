<!--
order: 3
-->

# To Improve

- For now, Slash and Jail take effect instantly at the end of block, but we will need to do slashing at the end of epoch and Jail instantly as it is.

- Didn't update `evidence` module so far which is related to Jail and we will need to check / modify.

```go
// Development plan
// — Implement changes for Msgs flow
// — — Add next epoch time calculation function and for completion time required messages, use that function
// — — For validator self undelegation, it could be required to do start on end blocker
// — — If delegation fail, just return money back to user at the time of epoch
// — — Implement TODOs on the PR #46
// — Implement CLI commands for querying
// — — BufferedValidators
// — — BufferedMsgCreateValidatorQueue, BufferedMsgEditValidatorQueue
// — — BufferedMsgUnjailQueue, BufferedMsgDelegateQueue, BufferedMsgRedelegationQueue, BufferedMsgUndelegateQueue
// — Fix existing tests / remove invalid tests
// — Write epoch related tests with new scenarios
// — — Simulation test is important for finding bugs [Ask Dev for questions) 
// — — Can easily add a simulator check to make sure all delegation amounts in queue add up to the same amount that’s in the EpochUnbondedPool
// — — I’d like it added as an invariant test for the simulator
// — — the simulator should check that the sum of all the queued delegations always equals the amount kept track in the data
// — — Staking/Slashing/Distribution module params are being modified by governance based on vote result instantly. We should test the effect.
// — — — Should test to see what would happen if max_validators is changed though, in the middle of an epoch
// — — we should define some new invariants that help check that everything is working smoothly with these new changes for 3 modules e.g. https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/keeper/invariants.go
// — — we should count all the delegation changes that happen during the epoch, and then make sure that the resulting change at the end of the epoch is actually correct
// — — If the validator that I delegated to double signs at block 16, I should still get slashed instantly because even though I asked to unbond at 14, they still used my power at block 16, I should only be not liable for slashes once my power is stopped being used
// — — On the converse of this, I should still be getting rewards while my power is being used.  I shouldn’t stop receiving rewards until block 20
```