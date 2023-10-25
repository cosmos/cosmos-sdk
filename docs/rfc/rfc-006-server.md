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

> Note: this example is coupled to CometBFT, but the idea is to allow for custom implementations of the server.

```go
type CometServer struct {
  // can load modules with either grpc, wasm, ffi or native. 
  // Depinject helps wire different configs
  // loaded from a config file that can compose different versions of apps
  // allows to sync from genesis with different config files
  // handles message execution 
  AppManager app.Manager
  // starts, stops and interacts with the consensus engine
  Comet comet.Server
  // mempool defines an application based mempool
  Mempool mempool.Mempool
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

The AppManager is responsible for loading the application and managing the application. The AppManager will be responsible for loading the application from a config file, loading the application from a genesis file, and managing the application. The AppManager will also be responsible for loading modules into the application.

The AppManager is responsible for state execution after block inclusion or during a predefined step in consensus. Today there are two methods of executing state, Optimistic Execution & Delayed Execution, with Comet. In the future if a new execution method is developed the AppManager will be responsible for executing state in that manner.

##### Transaction Hooks

When a transaction is being executed the AppManager will provide a per transaction before and after hook. This will replace the current antehandler and posthandler methods of a transaction. The replacement allows transaction validation to be run in a parallel manner. 

For example a module that would like to register a prehook on a message send would do so like this:

```go
func (m Module) HookMessages(hookRegistry) {
     sdk.RegisterPreHook(hookRegistry, func(ctx context.Context, msg bank.MsgSend) error {
          // business logic
     }
)}
```


The AppManager is responsible for:

* Loading the application from a config file
* Loading the application state from a genesis file
* Executing state after block inclusion or during a predefined step in consensus
* Upgrades of the application
* Syncing from genesis with a single binary that has many app configs loaded
* Querying state from any height


##### Upgrades

The Appmanager will be responsible for handling upgrades. Today in the Cosmos SDK we have an upgrade method of node operators must be present at the time of the upgrade or use a external process (cosmovisor) to swap the binary at the upgrade height. With the AppManager the design will be to allow for upgrades to be done in a rolling manner. A rolling upgrade is done by upgradeing the binary ahead of the upgrade height. Doing this allows advanced features like syncing from genesis with a single binary that has many app configs loaded and operatoring an archive node that allows users to request state from any height.

#### Comet

In this example we are using CometBFT, but the idea is to allow for custom implementations of the consensus engine. The Comet interface will be responsible for starting, stopping, and interacting with the consensus engine. The Comet interface will also be responsible for connecting the consensus engine to the application.

The responsability of the consensus engine is defined by the consensus engine, in our example the consensus engine componenet is responsible for: 

* Serving and receiving snapshots chunks
* Check transaction validity (CheckTX)
* Providing Comet with transactions to be included in a block (Prepare Proposal)
* Checking the validity of a block (Process Proposal)
* Deliver transactions to the AppManager

#### Mempool

The Mempool is responsible for storing transactions that must be included in a block. These transactions would be entered into a block by the consensus engine. 

The mempool is responsible for checking transaction validity, storing transactions and removing them after being included in a block. The mempool is modular and can be set by the application developer to use a custom mempool.

```go
type Mempool interface {
  service.Service
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
  // Validation of the transaction can be done at insertion or before dissemination
	Insert(context.Context, [][]byte) error

	// Select returns a list of transations from the mempool. The total size of the bytes which can be included is included in order for the mempool to handle ordering. If txs are specified,
	// then they shall be incorporated into the returned txs.
	Select(context.Context, [][]byte, int32) [][]byte

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove transactions from the mempool, returning an error
	// upon failure.
	Remove([][]byte) error
}
```

#### Storage

The storage is responsible for storing the application state. The storage is modular and can be set by the application developer to use a custom storage. 

The interface is being defined by the Storage working group and will be included in this RFC once it is completed.

#### Transaction Validation

The transaction validation is responsible for checking the validity of a transaction. In the current design this is handled by the antehandler. There is a blend of responsabilities currently with the anteHandler. Some useres use it for transaction hooks, some use it for custom transation validation. The transaction validation here is only to be used for validating if a transaction is valid or not. Separating the validation from the execution path allows us to check tx validation in a asynchronous manner.

```go
type TxValidation interface {
  // ValidateBasicTx validates a transaction 
  ValidateTx([]byte) error
  // ValidateBasicTxs validates a batch of transactions
  ValidateTxAsync([][]byte) error
  // Cache is a map of txs that may have already been verified
  Cache() TxCache
}
```

#### Transaction Codec

The transaction codec is responsible for encoding and decoding transactions. The transaction codec is modular and can be set by the application developer to use a custom transaction codec.

```go 
type TxCodec[T any] interface {
  // Encode encodes a transaction
  Encode(T) ([]byte, error)
  // Decode decodes a transaction
  Decode([]byte) (T, error)
}
```

#### Service

Service Defines an interface that will run in the background and can be started and stopped.

```go
type Service interface {
  Start() error
  Stop() error
}
```


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
