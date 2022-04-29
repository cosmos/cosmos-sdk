# ADR 050: Auto-Compound Rewards

## Changelog

* April 17, 2022: Initial Draft

## Status

Draft (not implemented)

## Abstract

This ADR describes a modification of the `x/distribution` module's functionality
to allow users to request the ability to auto-compound their rewards to their
delegated validators on-chain.

## Context

As of SDK version v0.45.x, the `x/distribution` module defines a mechanism in
which delegators receive rewards by delegating voting power to validators in the
form of a native staking token. The reward distribution itself happens in a lazy
fashion and is defined by the [F1 specification](https://drops.dagstuhl.de/opus/volltexte/2020/11974/pdf/OASIcs-Tokenomics-2019-10.pdf).
In other words, delegators accumulate "unrealized" rewards having to explicitly
execute message(s) on-chain in order to withdraw said rewards. This provides the
ability for the chain to not have to explicitly distribute rewards to delegators
on a block-by-block basis which would otherwise make the network crawl to halt
as the number of delegations increases over time.

However, it has been shown that there is a strong desire to auto-compound
distribution rewards. As a result, there has been a proliferation of scripts, tooling
and clients off-chain to facilitate such a mechanism. However, these methods are
ad-hoc, often provide a cumbersome UX, and do not scale well to multiple networks.

Thus, we propose a mechanism to modify the `x/distribution` module to allow for
delegators to auto-compound their rewards on-chain.

## Decision

In order to facilitate auto-compounding rewards, we need the ability for delegators
to opt into having their rewards auto-compounded. There are numerous ways to approach
this, where a simple method would to introduce a new message type, `MsgAutoCompoundRewards`,
defined as follows:

```protobuf
message MsgAutoCompoundRewards {
  option (cosmos.msg.v1.signer) = "delegator_address";

  // ...

  string delegator_address     = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string src_validator_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string dst_validator_address = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  bool   enable                = 4;
}
```

Recall, `x/distribution` executes reward withdrawal via a tuple, (`DelegatorAddress`, `ValidatorAddress`).
Furthermore, a delegator can also define a different address to which the rewards
are withdrawn to. This means that we need to know the tuple to execute the withdraw
and the withdraw address in order to send the rewards from when auto-compounding.

To reflect this in `MsgAutoCompoundRewards`, the `delegator_address` and `src_validator_address`
fields act as the tuple to execute the withdraw. We can use `delegator_address`
to find the withdraw address. Finally, the `dst_validator_address` defines the
validator address to delegate the withdrawn rewards to, from the withdraw address.
We imagine in most instances, `src_validator_address` and `dst_validator_address`
will be the same.

----

When a delegator wants to have their "unrealized" rewards be withdrawn and
automatically delegated to the relative validator(s), they would broadcast a
`MsgAutoCompoundRewards` transaction with `enable` set to `true` and a
`validator_address` that would have the withdrawn rewards delegated to. To stop
or disable auto-compounding, the user would send the same transaction with
`enable` set to `false` (`validator_address` can be omitted in this case).

In addition, we require the `x/distribution` module to use an additional state
index to store the records for delegators. When a user submits a `MsgAutoCompoundRewards`
transaction with `enable` set to `true`, we store a record with the following
key and value:

```text
<prefixByte> | address.MustLengthPrefix(DelegatorAddress) -> validator_address
```

When a user decides to disable auto-compounding rewards by setting `enable` to
`false`, we delete the record stored under the above key.

Given that we now have such a key ordering in state, we can iterate over the all
the relevant records using the dedicated `<prefixByte>` using a prefix `KVStore`,
which will allow us to withdraw rewards and re-delegate to validators. This can
be defined as follows in `x/distribution`:

```go
func (k Keeper) AutoCompoundRewards(ctx sdk.Context) {
  
}
```

## Consequences

### Backwards Compatibility

The changes proposed are naturally not backwards compatible with existing Cosmos
SDK versions as we introduce new message types and state additions along with the
requirement of using epoch-based actions.

### Positive

<!-- {positive consequences} -->

### Negative

<!-- {negative consequences} -->

### Neutral

<!-- {neutral consequences} -->

## Further Discussions

<!-- While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR. -->

## References

* [F1 Specification](https://drops.dagstuhl.de/opus/volltexte/2020/11974/pdf/OASIcs-Tokenomics-2019-10.pdf)
