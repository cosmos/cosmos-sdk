# RFC 006: Server

## Changelog

* October 18, 2023: Created

## Background

The Cosmos SDK is one of the most used frameworks to build a blockchain in the past years. While this is an achievement, there are more advanced users emerging (Berachain, Celestia, Rollkit, etc..) that require modifying the Cosmos SDK beyond the capabilities of the current framework. Within this RFC we will walk through the current pitfalls and proposed modifications to the Cosmos SDK to allow for more advanced users to build on top of the Cosmos SDK. 

Currently, the Cosmos SDK is tightly coupled with CometBFT in both production and in testing, with more environments emerging offering a simple and efficient manner to modify the Cosmos SDK to take advantage of these environments is necessary. Today, users must fork and maintain baseapp in order to modify the Cosmos SDK to work with these environments. This is not ideal as it requires users to maintain a fork of the Cosmos SDK and keep it up to date with the latest changes. We have seen this cause issues and forces teams to maintain a small team of developers to maintain the fork.

Secondly the current design, while it works, can have edge cases. With the combination of transaction validation, message execution and interaction with the consensus engine, it can be difficult to understand the flow of the Cosmos SDK. This is especially true when trying to modify the Cosmos SDK to work with a new consensus engine. Some of these newer engines also may want to modify ABCI or introduce a custom interface to allow for more advanced features, currently this is not possible unless you fork both CometBFT and the Cosmos SDK.

> The next section is the "Background" section. This section should be at least two paragraphs and can take up to a whole 
> page in some cases. The guiding goal of the background section is: as a newcomer to this project (new employee, team 
> transfer), can I read the background section and follow any links to get the full context of why this change is  
> necessary? 
> 
> If you can't show a random engineer the background section and have them acquire nearly full context on the necessity 
> for the RFC, then the background section is not full enough. To help achieve this, link to prior RFCs, discussions, and 
> more here as necessary to provide context so you don't have to simply repeat yourself.


## Proposal

The proposal is to allow users to create custom server implementations that can reuse existing features but also allow custom implementations. 

### Server

The server is the main entry point for the Cosmos SDK. It is responsible for starting the application, initializing the application, and starting the application. The server is also responsible for starting the consensus engine and connecting the consensus engine to the application. Each consensus engine will have a custom server implementation that will be responsible for starting the consensus engine and connecting it to the application.

While there will be default implementations provided by the Cosmos SDK if an application like Evmos or Berchain would like to implement their own server they can. This will allow for more advanced features to be implemented and allow for more advanced users to build on top of the Cosmos SDK.

```go
func NewGrpcCometServer(..) {..}
func NewGrpcRollkitServer(..) {..}
func NewEvmosCometServer(..) {..}
func NewPolarisCometServer(..) {..}
```

A server will consist of the following components, but is not limited to the ones included here. 

```go
type CometServer struct {
  // can load modules with either grpc, wasm, ffi or native. 
  // Depinject helps wire different configs
  // loaded from a config file that can compose different versions of apps
  // allows to sync from genesis with different config files
  // handles message execution 
  AppManager app.Manager
  // starts, stops and interacts with the consensus engine
  Consensus consensus.Engine
  // manages storage of application state
  Store store.RootStore 
  // manages application state snapshots
  StateSync snapshot.Manager 
  // transaction validation
  TxValidation core.TxValidation
  // decoder for trancations
  TxCodec core.TxCodec 
}
```

#### AppManager

#### Consensus

#### Storage

#### Transaction Validation

#### Transaction Codec


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
