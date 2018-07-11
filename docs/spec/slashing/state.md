## State

### Signing Info

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
- SigningBitArray: ` 0x02 | ValTendermintAddr | LittleEndianUint64(signArrayIndex) -> VarInt(didSign)`

The first map allows us to easily lookup the recent signing info for a
validator, according to the Tendermint validator address. The second map acts as
a bit-array of size `SIGNED_BLOCKS_WINDOW` that tells us if the validator signed for a given index in the bit-array.

The index in the bit-array is given as little endian uint64.

The result is a `varint` that takes on `0` or `1`, where `0` indicates the
validator did not sign the corresponding block, and `1` indicates they did.

Note that the SigningBitArray is not explicitly initialized up-front. Keys are
added as we progress through the first `SIGNED_BLOCKS_WINDOW` blocks for a newly
bonded validator.

The information stored for tracking validator liveness is as follows:

```go
type ValidatorSigningInfo struct {
  StartHeight           int64
  IndexOffset           int64
  JailedUntil           int64
  SignedBlocksCounter   int64
}

```

Where:
* `StartHeight` is set to the height that the candidate became an active validator (with non-zero voting power).
* `IndexOffset` is incremented each time the candidate was a bonded validator in a block (and may have signed a precommit or not).
* `JailedUntil` is set whenever the candidate is revoked due to downtime
* `SignedBlocksCounter` is a counter kept to avoid unnecessary array reads. `SignedBlocksBitArray.Sum() == SignedBlocksCounter` always.
