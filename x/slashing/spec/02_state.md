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
power included in the `LastCommitInfo` and +2/3 (see [fee distribution](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/_proposals/README.md)).

```
type LastCommitInfo struct {
	Round int32
	Votes []VoteInfo
}
```

Validators are penalized for failing to be included in the `LastCommitInfo` for some
number of blocks by being automatically jailed, potentially slashed, and unbonded.

Information about validator's liveness activity is tracked through `ValidatorSigningInfo`.
It is indexed in the store as follows:

- ValidatorSigningInfo: ` 0x01 | ConsAddress -> ProtocolBuffer(ValSigningInfo)`
- MissedBlocksBitArray: ` 0x02 | ConsAddress | LittleEndianUint64(signArrayIndex) -> VarInt(didMiss)`  (varint is a number encoding format)

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

```protobuf
// ValidatorSigningInfo defines a validator's signing info for monitoring their
// liveness activity.
message ValidatorSigningInfo {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = false;

  string address = 1;
  // Height at which validator was first a candidate OR was unjailed
  int64 start_height = 2;
  // Index which is incremented each time the validator was a bonded
  // in a block and may have signed a precommit or not. This in conjunction with the
  // `SignedBlocksWindow` param determines the index in the `MissedBlocksBitArray`.
  int64 index_offset = 3;
  // Timestamp until which the validator is jailed due to liveness downtime.
  google.protobuf.Timestamp jailed_until = 4;
  // Whether or not a validator has been tombstoned (killed out of validator set). It is set
  // once the validator commits an equivocation or for any other configured misbehiavor.
  bool tombstoned = 5;
  // A counter kept to avoid unnecessary array reads.
  // Note that `Sum(MissedBlocksBitArray)` always equals `MissedBlocksCounter`.
  int64 missed_blocks_counter = 6 [(gogoproto.moretags);
}
```
