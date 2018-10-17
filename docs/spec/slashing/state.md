# State

## Signing Info

Every block includes a set of precommits by the validators for the previous block, 
known as the LastCommit. A LastCommit is valid so long as it contains precommits from +2/3 of voting power.

Proposers are incentivized to include precommits from all
validators in the LastCommit by receiving additional fees
proportional to the difference between the voting power included in the
LastCommit and +2/3 (see [TODO](https://github.com/cosmos/cosmos-sdk/issues/967)).

Validators are penalized for failing to be included in the LastCommit for some
number of blocks by being automatically unbonded.

Information about validator activity is tracked in a `ValidatorSigningInfo`. 
It is indexed in the store as follows:

- SigningInfo: ` 0x01 | ValTendermintAddr -> amino(valSigningInfo)`
- MissedBlocksBitArray: ` 0x02 | ValTendermintAddr | LittleEndianUint64(signArrayIndex) -> VarInt(didMiss)`

The first map allows us to easily lookup the recent signing info for a
validator, according to the Tendermint validator address. The second map acts as
a bit-array of size `SIGNED_BLOCKS_WINDOW` that tells us if the validator missed the block for a given index in the bit-array.

The index in the bit-array is given as little endian uint64.

The result is a `varint` that takes on `0` or `1`, where `0` indicates the
validator did not miss (did sign) the corresponding block, and `1` indicates they missed the block (did not sign).

Note that the MissedBlocksBitArray is not explicitly initialized up-front. Keys are
added as we progress through the first `SIGNED_BLOCKS_WINDOW` blocks for a newly
bonded validator.

The information stored for tracking validator liveness is as follows:

```go
type ValidatorSigningInfo struct {
    StartHeight           int64     // Height at which the validator became able to sign blocks
    IndexOffset           int64     // Offset into the signed block bit array
    JailedUntilHeight     int64     // Block height until which the validator is jailed,
                                    // or sentinel value of 0 for not jailed
    MissedBlocksCounter   int64     // Running counter of missed blocks
}

```

Where:
* `StartHeight` is set to the height that the candidate became an active validator (with non-zero voting power).
* `IndexOffset` is incremented each time the candidate was a bonded validator in a block (and may have signed a precommit or not).
* `JailedUntil` is set whenever the candidate is jailed due to downtime
* `MissedBlocksCounter` is a counter kept to avoid unnecessary array reads. `MissedBlocksBitArray.Sum() == MissedBlocksCounter` always.

## Slashing Period

A slashing period is a start and end block height associated with a particular validator,
within which only the "worst infraction counts" (see the [Overview](overview.md)): the total
amount of slashing for infractions committed within the period (and discovered whenever) is
capped at the penalty for the worst offense.

This period starts when a validator is first bonded and ends when a validator is slashed & jailed
for any reason. When the validator rejoins the validator set (perhaps through unjailing themselves,
and perhaps also changing signing keys), they enter into a new period.

Slashing periods are indexed in the store as follows:

- SlashingPeriod: ` 0x03 | ValTendermintAddr | StartHeight -> amino(slashingPeriod) `

This allows us to look up slashing period by a validator's address, the only lookup necessary,
and iterate over start height to efficiently retrieve the most recent slashing period(s)
or those beginning after a given height.

```go
type SlashingPeriod struct {
    ValidatorAddr         sdk.ValAddress      // Tendermint address of the validator
    StartHeight           int64               // Block height at which slashing period begin
    EndHeight             int64               // Block height at which slashing period ended
    SlashedSoFar          sdk.Rat             // Fraction slashed so far, cumulative
}
```
