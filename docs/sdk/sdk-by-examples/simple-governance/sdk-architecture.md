## Architecture of a SDK-app

The Cosmos-SDK gives the basic template for a Tendermint-based blockchain application. You can find this template [here](https://github.com/cosmos/cosmos-sdk).

In essence, a blockchain application is simply a replicated state-machine. There is a state (e.g. for a cryptocurrency, how many coins each account holds), and transactions that trigger a state transition. As the application developer you define the state, the transaction types and how different transactions modify the state.

### Modularity

The Cosmos-SDK is a module-based framework. Each module is in itself a little state-machine that can be easily combined with other modules to produce a coherent application. In other words, modules define a sub-section of the global state and of the transaction types. Then, it is the job of the root application to route transactions to the correct modules depending on their respective types. To understand this process, let us take a look at a simplified standard cycle of the state-machine.

Upon receiving a transaction from the Tendermint Core engine, here is what the *Application* does:

1. Decode the transaction and get the message
2. Route the message to the appropriate module using the `Msg.Type()` method
3. Run the transaction in the module. Modify the state if the transaction is valid.
4. Return new state or error message

Steps 1, 2 and 4 are handled by the root application. Step 3 is handled by the appropriate module.

### SDK Components

With this in mind, let us go through the important directories of the SDK:

- `baseapp`: This defines the template for a basic application. Basically it implements the ABCI protocol so that your Cosmos-SDK application can communicate with the underlying Tendermint node.
- `client`: Command-Line Interface to interact with the application
- `server`: REST server to communicate with the node
- `examples`: Contains example on how to build a working application based on `baseapp` and modules
- `store`: Contains code for the multistore. The multistore is basically your state. Each module can create any number of KVStores from the multistore. Be careful to properly handle access rights to each store with appropriate `keepers`.
- `types`: Common types required in any SDK-based application.
- `x`: This is where modules live. You will find all the already-built modules in this directory. To use any of these modules, you just need to properly import them in your application. We will see how in the [App - Bridging it all together] section.

### Introductory Coderun

#### KVStore

The KVStore provides the basic persistence layer for your SDK application.

```go
type KVStore interface {
    Store

    // Get returns nil iff key doesn't exist. Panics on nil key.
    Get(key []byte) []byte

    // Has checks if a key exists. Panics on nil key.
    Has(key []byte) bool

    // Set sets the key. Panics on nil key.
    Set(key, value []byte)

    // Delete deletes the key. Panics on nil key.
    Delete(key []byte)

    // Iterator over a domain of keys in ascending order. End is exclusive.
    // Start must be less than end, or the Iterator is invalid.
    // CONTRACT: No writes may happen within a domain while an iterator exists over it.
    Iterator(start, end []byte) Iterator

    // Iterator over a domain of keys in descending order. End is exclusive.
    // Start must be greater than end, or the Iterator is invalid.
    // CONTRACT: No writes may happen within a domain while an iterator exists over it.
    ReverseIterator(start, end []byte) Iterator

    // TODO Not yet implemented.
    // CreateSubKVStore(key *storeKey) (KVStore, error)

    // TODO Not yet implemented.
    // GetSubKVStore(key *storeKey) KVStore
 }
```

You can mount multiple KVStores onto your application, e.g. one for staking, one for accounts, one for IBC, and so on.

```go
 app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake, app.keySlashing)
```

The implementation of a KVStore is responsible for providing any Merkle proofs for each query, if requested.

```go
 func (st *iavlStore) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
```

Stores can be cache-wrapped to provide transactions at the persistence level (and this is well supported for iterators as well). This feature is used to provide a layer of transactional isolation for transaction processing after the "AnteHandler" deducts any associated fees for the transaction. Cache-wrapping can also be useful when implementing a virtual-machine or scripting environment for the blockchain.

#### go-amino

The Cosmos-SDK uses [go-amino](https://github.com/cosmos/cosmos-sdk/blob/96451b55fff107511a65bf930b81fb12bed133a1/examples/basecoin/app/app.go#L97-L111) extensively to serialize and deserialize Go types into Protobuf3 compatible bytes.

Go-amino (e.g. over https://github.com/golang/protobuf) uses reflection to encode/decode any Go object.  This lets the SDK developer focus on defining data structures in Go without the need to maintain a separate schema for Proto3. In addition, Amino extends Proto3 with native support for interfaces and concrete types.

For example, the Cosmos-SDK's `x/auth` package imports the PubKey interface from `tendermint/go-crypto` , where PubKey implementations include those for _Ed25519_ and _Secp256k1_.  Each `auth.BaseAccount` has a PubKey.

```go
 // BaseAccount - base account structure.
 // Extend this by embedding this in your AppAccount.
 // See the examples/basecoin/types/account.go for an example.
 type BaseAccount struct {
    Address  sdk.Address   `json:"address"`
    Coins    sdk.Coins     `json:"coins"`
    PubKey   crypto.PubKey `json:"public_key"`
    Sequence int64         `json:"sequence"`
 }
```

Amino knows what concrete type to decode for each interface value based on what concretes are registered for the interface.

For example, the `Basecoin` example app knows about _Ed25519_ and _Secp256k1_ keys because they are registered by the app's `codec` below:

```go
wire.RegisterCrypto(cdc) // Register crypto.
```

For more information on Go-Amino, see https://github.com/tendermint/go-amino.

#### Keys, Keepers, and Mappers

The Cosmos-SDK is designed to enable an ecosystem of libraries that can be imported together to form a whole application. To make this ecosystem more secure, we've developed a design pattern that follows the principle of least-authority.

`Mappers` and `Keepers` provide access to KV stores via the context. The only difference between the two is that a `Mapper` provides a lower-level API, so generally a `Keeper` might hold references to other Keepers and `Mappers` but not vice versa.

`Mappers` and `Keepers` don't hold any references to any stores directly.  They only hold a _key_ (the `sdk.StoreKey` below):

```go
type AccountMapper struct {

    // The (unexposed) key used to access the store from the Context.
    key sdk.StoreKey

    // The prototypical Account concrete type.
    proto Account

    // The wire codec for binary encoding/decoding of accounts.
    cdc *wire.Codec
 }
```

This way, you can hook everything up in your main `app.go` file and see what components have access to what stores and other components.

```go
// Define the accountMapper.
 app.accountMapper = auth.NewAccountMapper(
    cdc,
    app.keyAccount,      // target store
    &types.AppAccount{}, // prototype
 )
```

Later during the execution of a transaction (e.g. via ABCI `DeliverTx` after a block commit) the context is passed in as the first argument.  The context includes references to any relevant KV stores, but you can only access them if you hold the associated key.

```go
 // Implements sdk.AccountMapper.
 func (am AccountMapper) GetAccount(ctx sdk.Context, addr sdk.Address) Account {
    store := ctx.KVStore(am.key)
    bz := store.Get(addr)
    if bz == nil {
        return nil
    }
    acc := am.decodeAccount(bz)
    return acc
 }
```

`Mappers` and `Keepers` cannot hold direct references to stores because the store is not known at app initialization time.  The store is dynamically created (and wrapped via `CacheKVStore` as needed to provide a transactional context) for every committed transaction (via ABCI `DeliverTx`) and mempool check transaction (via ABCI `CheckTx`).

#### Tx, Msg, Handler, and AnteHandler

A transaction (`Tx` interface) is a signed/authenticated message (`Msg` interface).

Transactions that are discovered by the Tendermint mempool are processed by the `AnteHandler` (_ante_ just means before) where the validity of the transaction is checked and any fees are collected.

Transactions that get committed in a block first get processed through the `AnteHandler`, and if the transaction is valid after fees are deducted, they are processed through the appropriate Handler.

In either case, the transaction bytes must first be parsed. The default transaction parser uses Amino. Most SDK developers will want to use the standard transaction structure defined in the `x/auth` package (and the corresponding `AnteHandler` implementation also provided in `x/auth`):

```go
 // StdTx is a standard way to wrap a Msg with Fee and Signatures.
 // NOTE: the first signature is the FeePayer (Signatures must not be nil).
 type StdTx struct {
    Msg        sdk.Msg        `json:"msg"`
    Fee        StdFee         `json:"fee"`
    Signatures []StdSignature `json:"signatures"`
 }
```

Various packages generally define their own message types.  The `Basecoin` example app includes multiple message types that are registered in `app.go`:

```go
sdk.RegisterWire(cdc)    // Register Msgs
 bank.RegisterWire(cdc)
 stake.RegisterWire(cdc)
 slashing.RegisterWire(cdc)
 ibc.RegisterWire(cdc)
```

Finally, handlers are added to the router in your `app.go` file to map messages to their corresponding handlers. (In the future we will provide more routing features to enable pattern matching for more flexibility).

```go
 // register message routes
 app.Router().
    AddRoute("auth", auth.NewHandler(app.accountMapper)).
    AddRoute("bank", bank.NewHandler(app.coinKeeper)).
    AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
    AddRoute("stake", stake.NewHandler(app.stakeKeeper))
```

#### EndBlocker

The `EndBlocker` hook allows us to register callback logic to be performed at the end of each block.  This lets us process background events, such as processing validator inflationary Atom provisions:

```go
// Process Validator Provisions
 blockTime := ctx.BlockHeader().Time // XXX assuming in seconds, confirm
 if pool.InflationLastTime+blockTime >= 3600 {
    pool.InflationLastTime = blockTime
    pool = k.processProvisions(ctx)
 }
```

By the way, the SDK provides a [staking module](https://github.com/cosmos/cosmos-sdk/tree/develop/x/stake), which provides all the bonding/unbonding funcionality for the Cosmos Hub.