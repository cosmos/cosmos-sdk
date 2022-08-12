<!--
order: 2
-->

# State

## Signing Info (Liveness)

Every block includes a set of precommits by the validators for the previous block,
known as the `LastCommitInfo` provided by Tendermint. A `LastCommitInfo` is valid so
long as it contains precommits from +2/3 of total voting power.

Proposers are incentivized to include precommits from all validators in the Tendermint `LastCommitInfo`
by receiving additional fees proportional to the difference between the voting
power included in the `LastCommitInfo` and +2/3 (see [fee distribution](x/distribution/spec/03_begin_block.md)).

```go
type LastCommitInfo struct {
	Round int32
	Votes []VoteInfo
}
```

Validators are penalized for failing to be included in the `LastCommitInfo` for some
number of blocks by being automatically jailed, potentially slashed, and unbonded.

Information about validator's liveness activity is tracked through `ValidatorSigningInfo`.
It is indexed in the store as follows:

* ValidatorSigningInfo: `0x01 | ConsAddrLen (1 byte) | ConsAddress -> ProtocolBuffer(ValSigningInfo)`
* MissedBlocksBitArray: `0x02 | ConsAddrLen (1 byte) | ConsAddress | LittleEndianUint64(signArrayIndex) -> VarInt(didMiss)` (varint is a number encoding format)

The first mapping allows us to easily lookup the recent signing info for a
validator based on the validator's consensus address.

The second mapping (`MissedBlocksBitArray`) acts
as a bit-array of size `SignedBlocksWindow` that tells us if the validator missed
the block for a given index in the bit-array. The index in the bit-array is given
as little endian uint64.
The result is a `varint` that takes on `0` or `1`, where `0` indicates the
validator did not miss (did sign) the corresponding block, and `1` indicates
they missed the block (did not sign).

Note that the `MissedBlocksBitArray` is not explicitly initialized up-front. Keys
are added as we progress through the first `SignedBlocksWindow` blocks for a newly
bonded validator. The `SignedBlocksWindow` parameter defines the size
(number of blocks) of the sliding window used to track validator liveness.

The information stored for tracking validator liveness is as follows:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/proto/cosmos/slashing/v1beta1/slashing.proto#L11-L33
