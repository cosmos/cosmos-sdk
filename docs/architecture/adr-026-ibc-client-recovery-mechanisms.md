# ADR 026: IBC Client Recovery Mechanisms

## Changelog

- 2020/06/23: Initial version
- 2020/08/06: Revisions per review & to reference version
- 2021/01/15: Revision to support substitute clients for unfreezing

## Status

*Accepted*

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
1. Require Tendermint light clients (ICS 07) & solo machine clients (ICS 06) to be created with the following additional flags
    1. `allow_governance_override_after_misbehaviour` (boolean, default false)
1. Require Tendermint light clients (ICS 07) to expose the following additional state mutation functions
    1. `Unfreeze()`, which unfreezes a light client after misbehaviour and clears any frozen height previously set
1. Add a new governance proposal type, `ClientUpdateProposal`, in the `x/ibc` module
    1. Extend the base `Proposal` with two client identifiers (`string`) and an initial height ('exported.Height'). 
    1. The first client identifier is the proposed client to be updated. This client must be either frozen or expired.
    1. The second client is a substitute client. It carries all the state for the client which may be updated. It must have identitical client and chain parameters to the client which may be updated (except for latest height, frozen height, and chain-id). It should be continually updated during the voting period. 
    1. The initial height represents the starting height consensus states which will be copied from the substitute client to the frozen/expired client.
    1. If this governance proposal passes, the client on trial will be updated with all the state of the substitute, if and only if:
        1. `allow_governance_override_after_expiry` is true and the client has expired (`Expired()` returns true)
        1. `allow_governance_override_after_misbehaviour` is true and the client has been frozen (`Frozen()` returns true)
            1. In this case, additionally, the client is unfrozen by calling `Unfreeze()`


Note that clients frozen due to misbehaviour must wait for the evidence to expire to avoid becoming refrozen. 

This ADR does not address planned upgrades, which are handled separately as per the [specification](https://github.com/cosmos/ibc/tree/master/spec/client/ics-007-tendermint-client#upgrades).

## Consequences

### Positive

- Establishes a mechanism for client recovery in the case of expiry
- Establishes a mechanism for client recovery in the case of misbehaviour
- Clients can elect to disallow this recovery mechanism if they do not wish to allow for it
- Constructing an ClientUpdate Proposal is as difficult as creating a new client

### Negative

- Additional complexity in client creation which must be understood by the user
- Coping state of the substitute adds complexity
- Governance participants must vote on a substitute client

### Neutral

No neutral consequences.

## References

- [Prior discussion](https://github.com/cosmos/ics/issues/421)
- [Epoch number discussion](https://github.com/cosmos/ics/issues/439)
- [Upgrade plan discussion](https://github.com/cosmos/ics/issues/445)
