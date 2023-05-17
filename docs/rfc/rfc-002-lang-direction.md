# RFC 002: Cross Language Direction

## Changelog

* 17-05-2023: Proposed

## Background

This document aims to propose a scope of work that introduces support for different programming languages within the Cosmos SDK. Unlike traditional RFCs, it will also provide a high-level overview of why this proposal is being made and how it will impact users.

Currently, the Cosmos SDK allows developers to write modules in Golang, which has been a popular choice. However, the broader Cosmos ecosystem has witnessed the widespread adoption of other languages through virtual machines (VMs) such as Cosmwasm (Wasm), Agoric (JS), Polaris (Solidity), and Ethermint (Solidity). Additionally, within the wider blockchain ecosystem, Rust and various Rust subset VMs have gained significant traction.

Given this landscape, the question arises: should the Cosmos SDK support different languages and offer better support for their usage? To address this question, it's helpful to view the SDK as consisting of two distinct layers: the kernel space and the user space. In this perspective, modules represent the kernel space, while VMs constitute the user space. When discussing the support for different languages, both spaces need to be considered.

The primary objective is to enable modules to be written in different programming languages, allowing them to serve as core components of a node. Simultaneously, users should have the ability to write VMs in Rust and connect them to a general interface that provides similar, albeit limited, functionality as a module. This approach ensures flexibility in language choice for both core development and user customization within the Cosmos ecosystem.


## Proposal

> The next required section is "Proposal" or "Goal". Given the background above, this section proposes a solution. 
> This should be an overview of the "how" for the solution, but for details further sections will be used.


## Abandoned Ideas (Optional)

> As RFCs evolve, it is common that there are ideas that are abandoned. Rather than simply deleting them from the 
> document, you should try to organize them into sections that make it clear they're abandoned while explaining why they 
> were abandoned.
> 
> When sharing your RFC with others or having someone look back on your RFC in the future, it is common to walk the same 
> path and fall into the same pitfalls that we've since matured from. Abandoned ideas are a way to recognize that path 
> and explain the pitfalls and why they were abandoned.

## Descision

> This section describes alternative designs to the chosen design. This section
> is important and if an adr does not have any alternatives then it should be
> considered that the ADR was not thought through. 

## Consequences (optional)

> This section describes the resulting context, after applying the decision. All
> consequences should be listed here, not just the "positive" ones. A particular
> decision may have positive, negative, and neutral consequences, but all of them
> affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section
> describing these incompatibilities and their severity. The ADR must explain
> how the author proposes to deal with these incompatibilities. ADR submissions
> without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

> {positive consequences}

### Negative

> {negative consequences}

### Neutral

> {neutral consequences}



### References

> Links to external materials needed to follow the discussion may be added here.
>
> In addition, if the discussion in a request for comments leads to any design
> decisions, it may be helpful to add links to the ADR documents here after the
> discussion has settled.

## Discussion

> This section contains the core of the discussion.
>
> There is no fixed format for this section, but ideally changes to this
> section should be updated before merging to reflect any discussion that took
> place on the PR that made those changes.
