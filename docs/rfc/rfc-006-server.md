# RFC 006: Server

## Changelog

* October 18, 2023: Created

## Background

The Cosmos SDK is one of the most used frameworks to build a blockchain in the past years. While this is an achievement, there are more advanced users emerging (Berachain, Celestia, Rollkit, etc..) that require modifying the Cosmos SDK beyond the capabilities of the current framework. Within this RFC we will walk through the current pitfalls and proposed modifications to the Cosmos SDK to allow for more advanced users to build on top of the Cosmos SDK. 

Currently, the Cosmos SDK is tightly coupled with CometBFT in both production and in testing, with more environments emerging offering a simple and efficient manner to modify the Cosmos SDK to take advantage of these environments is necessary. Today, users must fork and maintain baseapp in order to modify the Cosmos SDK to work with these environments. This is not ideal as it requires users to maintain a fork of the Cosmos SDK and keep it up to date with the latest changes. We have seen this cause issues and forces teams to maintain a small team of developers to maintain the fork.

Secondly the current design, while it works, can have edge cases. With the combination of transaction validation, message execution and interaction with the consensus engine, it can be difficult to understand the flow of the transaction execution. This is especially true when trying to modify the Cosmos SDK to work with a new consensus engine. Some of these newer engines also may want to modify ABCI or introduce a custom interface to allow for more advanced features, currently this is not possible unless you fork both CometBFT and the Cosmos SDK.


## Proposal

This proposal is centered around modularity and simplicity of the Cosmos SDK. The goal is to allow for more advanced users to build on top of the Cosmos SDK without having to maintain a fork of the Cosmos SDK. Within the design we create clear separations between state transition, application, mempool, client and consensus. These five unique and separate componets interact with each other through well defined interfaces. While we aim to create a generalized framework, we understand that we can not fit every use case into a few simple interfaces, this is why we opted to create the seperation between the five components. If a user would like to extend one componenet they are able to do so without, potentially, needing to fork other components. This brings in a new case in which users will not need to fork the entirety of the Cosmos SDK, but only a single component. 

### Server

The server is the workhorse of the state machine. It is where all the components are initialized, combined and started. The server is where the consensus engine lives, meaning that every server will be custom to a consensus engine or logic for the application. The default server will be using comet with its async client to enable better concurrency. 

#### Consensus

Consensus is part of server and the component that controls the rest of the state machine. It receives a block and tells the STF (state transition function) what to execute, in which order and possibly in parallel. The consensus engine receives multiple componenets from the server, application manager, mempool and a smaller component, the transaction codec. 



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
