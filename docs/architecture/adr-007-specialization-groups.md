# ADR 007: Specialization Groups

## Changelog

- 2019 Jul 31 : Initial Conception

## Context

This idea was first conceived of in order to fulfill the use case of the
creation of a decentralized Computer Emergency Response Team (CERT), Whose
members would be elected by the Cosmos Community and would fulfill the role of
coordinating the Cosmos community under an emergency situations. This thinking
can be further abstracted into the conception of "blockchain specialization
groups". 

The creation of these groups are the beginning of specialization capabilities
within a wider blockchain community which could be used to enable a certain
level of delegated responsibilities. Examples of specialization which could be
beneficial to a blockchain community include: code auditing, emergency response,
code development etc. This type of community organization paves the way for
individual stakeholders to delegate votes by issue type, if in the future
governance proposals include a field for issue type. 


## Decision

A specialization group can be broadly broken down into
the following functions: 
 - Membership Admittance
   - via governance
   - via special appointment
 - Membership Acceptance
 - Membership Revocation
   - (probably) Without Penalty 
     - member steps down (self-Revocation)
     - replaced by new member from governance
   - (probably) With Penalty 
     - due to breach of soft-agreement (determined through governance)
     - due to breach of hard-agreement (determined by code) 
 - Execution of Duties
   - Special transactions which only execute for members of a specialization group
 - Compensation
   - Group compensation (further distribution decided by the specialization group) 
   - Individual compensation for all constituents of a group from the
     greater community

Election of the members of a specialization group can happen in a wide variety
of ways and be subject to an arbitrary number of associated rules. For its
initial implementation as a part of this ADR we recommend that a general
election abstraction (`Electionator`) is provided as well as a basic
implementation of that abstraction which allows for a continuous election of
members of a specialization group. 

``` golang
// The Electionator abstraction covers the concept space for 
// a wide variety of elections. 
type Electionator interface {
    
    // is the election object active
    Active() bool 

    // functionality to execute for when a member casts a vote, here the
    // vote field is anticipated to be marshalled into a vote type used 
    // by an election
    Vote(addr sdk.Address, vote []byte)  

    // here lies all functionality to authenticate and execute changes for
    // when a member accepts being elected
    Accept(sdk.Address) 

    // Register a revoker object
    RegisterRevoker(Revoker)

    // No more revokers may be registered after this function is called
    SealRevokers()

    // query for the current winner(s) of this election based on arbitrary
    // election ruleset
    QueryWinners() []sdk.Address 
}

// Revoker defines the function required for an membership revocation ruleset
// used by a specialazation group. This could be used to create self revoking,
// and evidence based revoking, etc. Revokers types may be created and
// reused for different election types. 
// 
// When revoking the "cause" bytes may be arbirarily marshalled into evidence,
// memos, etc.
type Revoker interface {
    RevokeName() string      // identifier for this revoker type 
    RevokeMember(revokeAddr sdk.Address, cause []byte) (successful bool)
}
```

Certain level of commonality likely exists between the existing code within
`x/governance` and required functionality of elections. This common
functionality should be abstracted during implementation time. 

The specialization group abstraction firstly depends on the `Electionator`
but also further defines the con

``` golang
type SpecializationGroup interface {
   election Election // election object
   p
   
}
```

TODO talk about throttlers 



## Status

> Proposed

## Consequences

### Positive

 - Increases specialization capabilities of a blockchain

### Negative

 - Could be used to negatively increase centralization within a community

### Neutral

## References
 
