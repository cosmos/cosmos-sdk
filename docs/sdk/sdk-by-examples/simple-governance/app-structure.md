## Application structure

Now, that we have built all the pieces we need, it is time to integrate them into the application. Let us exit the `/x` director go back at the root of the SDK directory.


```bash
// At root level of directory
cd app
```

We are ready to create our simple governance application!

*Note: You can check the full file (with comments!) [here](link)*

The `app.go` file is the main file that defines your application. In it, you will declare all the modules you need, their keepers, handlers, stores, etc. Let us take a look at each section of this file to see how the application is constructed.

Secondly, we need to define the name of our application.

```go
const (
    appName = "SimpleGovApp"
)
```

Then, let us define the structure of our application.

```go
// Extended ABCI application
type SimpleGovApp struct {
    *bam.BaseApp
    cdc *wire.Codec

    // keys to access the substores
    capKeyMainStore      *sdk.KVStoreKey
    capKeyAccountStore   *sdk.KVStoreKey
    capKeyStakingStore   *sdk.KVStoreKey
    capKeySimpleGovStore *sdk.KVStoreKey

    // keepers
    feeCollectionKeeper auth.FeeCollectionKeeper
    coinKeeper          bank.Keeper
    stakeKeeper         simplestake.Keeper
    simpleGovKeeper     simpleGov.Keeper

    // Manage getting and setting accounts
    accountMapper auth.AccountMapper
}
```

- Each application builds on top of the `BaseApp` template, hence the pointer.
- `cdc` is the codec used in our application.
- Then come the keys to the stores we need in our application. For our simple governance app, we need 3 stores + the main store.
- Then come the keepers and mappers.

Let us do a quick reminder so that it is  clear why we need these stores and keepers. Our application is primarily based on the `simple_governance` module. However, we have established in section [Keepers for our app](module-keeper.md) that our module needs access to two other modules: the `bank` module and the `stake` module. We also need the `auth` module for basic account functionalities. Finally, we need access to the main multistore to declare the stores of each of the module we use.