# ADR 008: Decentralized Computer Emergency Response Team (dCERT) Group

## Changelog

- 2019 Jul 31 : Initial Conception

## Context

In order to reduce the number of parties involved with handling sensitive
information in an emergency scenario, we propose the creation of a
specialization group named The Decentralized Computer Emergency Response Team
(dCERT).  Initially this group's role is intended to serve as coordinators
between the validators, bug-hunters, developers, and other members of the wider
community.  During a crisis time, the dCERT group would facilitate and relay
input from a variety of stakeholders to the developers who are actively
devising a patch to the software, this way sensitive information does not need
to be publicly disclosed while as some input from the community can still be
gained. 

Additionally, a special privilege is proposed for the dCERT group:
the capacity to "circuit-break" (aka. temporarily disable)  a particular
transaction path. 

In the future it is foreseeable that the community may wish to expand the roles
of dCERT with further responsibilities such as the capacity to "pre-approve" a
security update hard-fork on behalf of the community prior to a full community
wide vote whereby the sensitive information would be revealed prior to a
vulnerability being patched on the live network.  

## Decision

The dCERT group is proposed to include an implementation of a `SpecializationGroup`
as defined in [ADR 007](./adr-007-specialization-groups.md). This will include the 
implementation of: 
 - continuous voting
 - slashing due to breach of soft contract
 - revoking a member due to breach of soft contract
 - compensation stipend from the community pool or other means decided by
   governance

This system necessitates the following new parameters: 
 - blockly stipend allowance per dCERT member (suggested ~$10K/yr of atoms)
 - maximum number of dCERT members (suggested 7) 
 - required staked slashable tokens for each dCERT member (suggested $10K of atoms)
 - unbonding time for dCERT staked token (suggested 3 week)
 - quorum for suspending a particular member (suggested 5/7) 

These parameters are expected to be implemented through the param keeper such 
that governance may change them at any given point. 

### Continuous Voting Electionator

An `Electionator` object is to be implemented with continuous voting is to be
implemented with the following specifications:
 - All delegation addresses may submit votes at any point which updates their 
   preferred representation on the dCERT group. 
 - Preferred representation may be arbitrarily split between address (ex. 50%
   to John, 25% to Sally, 25% to Carol) 
 - Addresses which control the greatest amount of prefered-representation are
   eligible to join the dCERT group (up the _maximum number of dCERT members_)
 - In order for a new member to be added to the dCERT group they must 
   send a transaction accepting their admission at which point the validity of
   their admission is to be confirmed. 

### Staking/Slashing

All members of the dCERT group must stake tokens _specifically_ to maintain
eligibility as a dCERT member. This staking mechanism must be designed with
an unbonding period. Slash a particular dCERT member due to soft-contract breach
should be performed by governance on a per member basis based on the magnitude
of the breach. The process flow is anticipated to be that a dCERT member 
is suspended by the dCERT group prior to being slashed by governance. 

### dCERT membership transactions

Active dCERT members 
 - change of the description of the dCERT group
 - circuit break a transaction route
 - vote to suspend a dCERT member. 

## Status

> Proposed

## Consequences

### Positive

 - Potential to reduces the number of parties to coordinate with during an emergency 
   - Reduction in possibility of disclosing sensitive information to malicious parties

### Negative

 - Centralization risks

### Neutral

## References
 
  (Specialization Groups ADR)[./adr-007-specialization-groups.md]
