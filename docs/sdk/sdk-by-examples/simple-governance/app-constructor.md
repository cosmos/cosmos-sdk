## Application constructor

**File: [`app/app.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/app/app.go)**

Now, we need to define the constructor for our application.

```go
func NewSimpleGovApp(logger log.Logger, db dbm.DB) *SimpleGovApp
```

In this function, we will:

- Create the codec

```go
var cdc = MakeCodec()
```

- Instantiate our application. This includes creating the keys to access each of the substores.

```go
// Create your application object.
    var app = &SimpleGovApp{
        BaseApp:              bam.NewBaseApp(appName, cdc, logger, db),
        cdc:                  cdc,
        capKeyMainStore:      sdk.NewKVStoreKey("main"),
        capKeyAccountStore:   sdk.NewKVStoreKey("acc"),
        capKeyStakingStore:   sdk.NewKVStoreKey("stake"),
        capKeySimpleGovStore: sdk.NewKVStoreKey("simpleGov"),
    }
```

- Instantiate the keepers. Note that keepers generally need access to other module's keepers. In this case, make sure you only pass an instance of the keeper for the functionality that is needed. If a keeper only needs to read in another module's store, a read-only keeper should be passed to it.

```go
app.coinKeeper = bank.NewKeeper(app.accountMapper)
app.stakeKeeper = simplestake.NewKeeper(app.capKeyStakingStore, app.coinKeeper,app.RegisterCodespace(simplestake.DefaultCodespace))
app.simpleGovKeeper = simpleGov.NewKeeper(app.capKeySimpleGovStore, app.coinKeeper, app.stakeKeeper, app.RegisterCodespace(simpleGov.DefaultCodespace))
```

- Declare the handlers.

```go
app.Router().
        AddRoute("bank", bank.NewHandler(app.coinKeeper)).
        AddRoute("simplestake", simplestake.NewHandler(app.stakeKeeper)).
        AddRoute("simpleGov", simpleGov.NewHandler(app.simpleGovKeeper))
```

- Initialize the application.

```go
// Initialize BaseApp.
    app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore, app.capKeySimpleGovStore, app.capKeyStakingStore)
    app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
    err := app.LoadLatestVersion(app.capKeyMainStore)
    if err != nil {
        cmn.Exit(err.Error())
    }
    return app
```