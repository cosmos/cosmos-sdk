# ADR 026: IBC Client Recovery Mechanisms

## Changelog

- 2020/06/23: Initial version
- 2020/08/06: Revisions per review & to reference version

## Status

*Proposed*

## Context

### Summary

At launch, IBC will be a novel protocol, without an experienced user-base. At the protocol layer, it is not possible to distinguish between client expiry or misbehaviour due to genuine faults (Byzantine behavior) and client expiry or misbehaviour due to user mistakes (failing to update a client, or accidentally double-signing). In the base IBC protocol and ICS 20 fungible token transfer implementation, if a client can no longer be updated, funds in that channel will be permanently locked and can no longer be transferred. To the degree that it is safe to do so, it would be preferable to provide users with a recovery mechanism which can be utilised in these exceptional cases.

### Exceptional cases

The state of concern is where a client associated with connection(s) and channel(s) can no longer be updated. This can happen for several reasons:

1. The chain which the client is following has halted and is no longer producing blocks/headers, so no updates can be made to the client
1. The chain which the client is following has continued to operate, but no relayer has submitted a new header within the unbonding period, and the client has expired
    1. This could be due to real misbehaviour (intentional Byzantine behaviour) or merely a mistake by validators, but the client cannot distinguish these two cases
1. The chain which the client is following has experienced a misbehaviour event, and the client has been frozen & thus can no longer be updated

### Security model

Two-thirds of the validator set (the quorum for governance, module participation) can already sign arbitrary data, so allowing governance to manually force-update a client with a new header after a delay period does not substantially alter the security model.

## Decision

We elect not to deal with chains which have actually halted, which is necessarily Byzantine behaviour and in which case token recovery is not likely possible anyways (in-flight packets cannot be timed-out, but the relative impact of that is minor).

1. Require Tendermint light clients (ICS 07) to be created with the following additional flags
    1. `allow_governance_override_after_expiry` (boolean, default false)
1. Require Tendermint light clients (ICS 07) to expose the following additional internal query functions
    1. `Expired() boolean`, which returns whether or not the client has passed the trusting period since the last update (in which case no headers can be validated)
1. Require Tendermint light clients (ICS 07) to expose the following additional state mutation functions
    1. `Unfreeze()`, which unfreezes a light client after misbehaviour and clears any frozen height previously set
1. Require Tendermint light clients (ICS 07) & solo machine clients (ICS 06) to be created with the following additional flags
    1. `allow_governance_override_after_misbehaviour` (boolean, default false)
1. Add a new governance proposal type, `ClientUpdateProposal`, in the `x/ibc` module
    1. Extend the base `Proposal` with a client identifier (`string`) and a header (`bytes`, encoded in a client-type-specific format)
    1. If this governance proposal passes, the client is updated with the provided header, if and only if:
        1. `allow_governance_override_after_expiry` is true and the client has expired (`Expired()` returns true)
        1. `allow_governance_override_after_misbehaviour` is true and the client has been frozen (`Frozen()` returns true)
            1. In this case, additionally, the client is unfrozen by calling `Unfreeze()`

Note additionally that the header submitted by governance must be new enough that it will be possible to update the light client after the new header is inserted into the client state (which will only happen after the governance proposal has passed).

This ADR does not address planned upgrades, which are handled separately as per the [specification](https://github.com/cosmos/ics/tree/master/spec/ics-007-tendermint-client#upgrades).

## Consequences

### Positive

- Establishes a mechanism for client recovery in the case of expiry
- Establishes a mechanism for client recovery in the case of misbehaviour
- Clients can elect to disallow this recovery mechanism if they do not wish to allow for it

### Negative

- Additional complexity in client creation which must be understood by the user
- Governance participants must pick a new header, which is a bit different from their usual tasks

### Neutral

No neutral consequences.

## References

- [Prior discussion](https://github.com/cosmos/ics/issues/421)
- [Epoch number discussion](https://github.com/cosmos/ics/issues/439)
- [Upgrade plan discussion](https://github.com/cosmos/ics/issues/445)
