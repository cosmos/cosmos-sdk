# POS-to-POA Migration Example

Working example of migrating a Cosmos SDK chain from Proof-of-Stake to Proof-of-Authority using the POA module. Includes a transitional binary, upgrade handler, and an end-to-end system test.

## Structure

```
examples/migrate-from-pos/
├── app.go                              # Transitional SimApp (POS keepers + POA module)
├── upgrades.go                         # Wires the upgrade handler into the app
├── ante.go                             # AnteHandler with POA fee routing
├── sample_upgrades/
│   └── upgrade_handler.go              # POS → POA upgrade handler + helpers
└── simd/                               # Binary entry point
```

## How It Works

The transitional SimApp (`app.go`) includes the POA module alongside staking/distribution/slashing **keepers** (not modules). These keepers are not registered in the module manager — they exist solely so the upgrade handler can read pre-upgrade POS state.

The upgrade handler (`sample_upgrades/upgrade_handler.go`) executes these steps atomically:

1. **Snapshot** bonded validators as POA validators
2. **Withdraw** all distribution rewards and commissions
3. **Force-unbond** all delegations (returns tokens to delegators)
4. **Complete** in-flight unbonding delegations immediately
5. **Remove** all in-flight redelegations
6. **Drain** staking pool dust to the community pool
7. **Fail** active governance proposals and refund deposits
8. **Initialize** the POA module via `InitGenesis`

Step 1 must precede all teardown. Step 2 must precede step 3 — `Unbond()` fires `BeforeDelegationSharesModified` which calls into distribution, so rewards must be withdrawn first.

After the upgrade, the transitional binary runs the chain with the leftover POS keepers sitting inert. Replace it with the clean [POA SimApp](../../simapp/) at the next scheduled upgrade.

## Key Binary Changes

The transitional binary differs from a standard POS simapp:

- **Ante handler**: `WithFeeRecipientModule(poatypes.ModuleName)` routes fees to POA. The chain panics without this.
- **Gov keeper**: Rewired with `NewPOACalculateVoteResultsAndVotingPowerFn` and POA gov hooks (restricts voting to POA validators).
- **Store upgrades**: Adds `poatypes.StoreKey`. Does NOT delete staking/distribution/slashing stores — the upgrade handler needs to read them. Delete them in the subsequent clean-binary upgrade.

## System Test

The end-to-end test at `../../tests/systemtests/upgrade_test.go` (`TestPOStoPoaUpgrade`):

1. Starts a 4-validator POS chain
2. Creates pre-upgrade state: third-party delegation, in-flight unbonding, active governance proposal
3. Submits and passes an upgrade proposal
4. Swaps to the transitional POA binary at the upgrade height
5. Verifies post-upgrade:
   - POA validators match pre-upgrade bonded set
   - Delegator tokens returned (~50M of 50M minus fees)
   - Unbonding delegation completed, tokens returned
   - Active proposal failed/rejected, deposit refunded
   - Total supply preserved within 0.1% tolerance
   - POA governance works (submit, vote, pass a new proposal)
   - Fee distribution routes to POA validators
   - Bank transfers work

### Running

```bash
cd ../../tests/systemtests
make test-upgrade
```

This builds both the POS binary (standard SDK simapp) and the transitional POA binary, then runs the full upgrade flow.

## ICS Consumer Chains

The same upgrade handler steps apply with one difference: snapshot the CCV validator set via `consumerKeeper.GetAllCCValidator(ctx)` instead of the staking bonded set. Also delete the `ccv/consumer` store key and remove the consumer module and IBC middleware from app wiring. The IBC channel to the provider will timeout naturally.
