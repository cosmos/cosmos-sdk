# ADR 008: Decentralized Computer Emergency Response Team (dCERT) Group

## Changelog

- 2019 Jul 31: Initial Draft

## Context

In order to reduce the number of parties involved with handling sensitive
information in an emergency scenario, we propose the creation of a
specialization group named The Decentralized Computer Emergency Response Team
(dCERT).  Initially this group's role is intended to serve as coordinators
between the validators, bug-hunters, developers, and other members of the wider
community.  During a time of crisis, the dCERT group would aggregate and relay
input from a variety of stakeholders to the developers who are actively
devising a patch to the software, this way sensitive information does not need
to be publicly disclosed while some input from the community can still be
gained. 

Additionally, a special privilege is proposed for the dCERT group:
the capacity to "circuit-break" (aka. temporarily disable)  a particular
message path. 

In the future it is foreseeable that the community may wish to expand the roles
of dCERT with further responsibilities such as the capacity to "pre-approve" a
security update on behalf of the community prior to a full community
wide vote whereby the sensitive information would be revealed prior to a
vulnerability being patched on the live network.  

## Decision

The dCERT group is proposed to include an implementation of a `SpecializationGroup`
as defined in [ADR 007](./adr-007-specialization-groups.md). This will include the 
implementation of: 
 - continuous voting
 - slashing due to breach of soft contract
 - revoking a member due to breach of soft contract
 - emergency disband of the entire dCERT group (ex. for colluding maliciously) 
 - compensation stipend from the community pool or other means decided by
   governance

This system necessitates the following new parameters: 
 - blockly stipend allowance per dCERT member (suggested ~$10K/yr of atoms)
 - maximum number of dCERT members (suggested 7) 
 - required staked slashable tokens for each dCERT member (suggested $10K of atoms)
 - unbonding time for dCERT staked token (suggested 3 week)
 - quorum for suspending a particular member (suggested 5/7) 
 - proposal wager for disbanding the dCERT group (suggested ~$100000 USD worth of atoms)

These parameters are expected to be implemented through the param keeper such 
that governance may change them at any given point. 

### Continuous Voting Electionator

An `Electionator` object is to be implemented with continuous voting is to be
implemented with the following specifications:
 - All delegation addresses may submit votes at any point which updates their 
   preferred representation on the dCERT group. 
 - Preferred representation may be arbitrarily split between addresses (ex. 50%
   to John, 25% to Sally, 25% to Carol) 
 - Addresses which control the greatest amount of preferred-representation are
   eligible to join the dCERT group (up the _maximum number of dCERT members_)
   - In the split situation where the dCERT group is full but a vying candidate 
     has the same amount of vote as an existing dCERT member, the existing 
     member should maintain its position. 
 - In order for a new member to be added to the dCERT group they must 
   send a transaction accepting their admission at which point the validity of
   their admission is to be confirmed. 
   - A sequence number is assigned when a member is added to dCERT group. 
     If a member leaves the dCERT group and then enters back, a new sequence number
     is assigned.  
 - If the dCERT group is already full and new member is admitted, the existing
   dCERT member with the lowest amount of votes is kicked from the dCERT group.
   - In the split situation where two addresses with the smallest number of
     votes have the same number of votes, the address with the smallest sequence 
     number maintains its position.  

### Staking/Slashing

All members of the dCERT group must stake tokens _specifically_ to maintain
eligibility as a dCERT member. This staking mechanism must be designed with an
unbonding period. Slash a particular dCERT member due to soft-contract breach
should be performed by governance on a per member basis based on the magnitude
of the breach.  The process flow is anticipated to be that a dCERT member is
suspended by the dCERT group prior to being slashed by governance.  

Membership suspension by the dCERT group takes place through a voting procedure
by the dCERT group members. After this suspension has taken place, a governance
proposal to slash the dCERT member must be submitted, if the proposal is not
approved by the time the rescinding member has completed unbonding their
tokens, then the tokens are no longer staked and unable to be slashed. 

Additionally in the case of an emergency situation of a colluding and malicious
dCERT group, the community needs the capability to disband the entire dCERT
group and likely fully slash them. This could be achieved though a special new
proposal type (implemented as a general governance proposal) which would halt
the functionality of the dCERT group until the proposal was concluded. This
special proposal type would likely need to also have a fairly large wager which
could be slashed if the proposal creator was malicious. The reason a large
wager should be required is because as soon as the proposal is made, the
capability of the dCERT group to halt message routes is put on temporarily
suspended, meaning that a malicious actor who created such a proposal could
then potentially exploit a bug during this period of time, with no dCERT group
capable of shutting down the exploitable message routes. 

### dCERT membership transactions

Active dCERT members 
 - change of the description of the dCERT group
 - circuit break a message route
 - vote to suspend a dCERT member. 

Here circuit-breaking refers to the capability to disable a groups of messages,
This could for instance mean: "disable all staking-delegation messages", or
"disable all distribution messages". This could be accomplished by verifying
that the message route has not been "circuit-broken" at CheckTx time (in
`baseapp/baseapp.go`). 

"unbreaking" a circuit is anticipated only to occur during a hard fork upgrade
meaning that no capability to unbreak a message route on a live chain is
required. 

Note also, that if there was a problem with governance voting (for instance a
capability to vote many times) then governance would be broken and should be
halted with this mechanism, it would be then up to the validator set to
coordinate and hard-fork upgrade to a patched version of the software where
governance is re-enabled (and fixed). If the dCERT group abuses this privilege
they should all be severely slashed.

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
