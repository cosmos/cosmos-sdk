# From Module To Application

## Application structure

Now, that we have built all the pieces we need, it is time to integrate them into the application. Let us exit the `/x` director go back at the root of the SDK directory.


```bash
// At root level of directory
cd app
```

We are ready to create our simple governance application!

*Note: You can check the full file (with comments!) [here](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/app/app.go)*

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
    cdc *codec.Codec

    // keys to access the substores
    capKeyMainStore      *sdk.KVStoreKey
    capKeyAccountStore   *sdk.KVStoreKey
    capKeyStakingStore   *sdk.KVStoreKey
    capKeySimpleGovStore *sdk.KVStoreKey

    // keepers
    feeCollectionKeeper auth.FeeCollectionKeeper
    bankKeeper          bank.Keeper
    stakeKeeper         simplestake.Keeper
    simpleGovKeeper     simpleGov.Keeper

    // Manage getting and setting accounts
    accountKeeper auth.AccountKeeper
}
```

- Each application builds on top of the `BaseApp` template, hence the pointer.
- `cdc` is the codec used in our application.
- Then come the keys to the stores we need in our application. For our simple governance app, we need 3 stores + the main store.
- Then come the keepers and mappers.

Let us do a quick reminder so that it is  clear why we need these stores and keepers. Our application is primarily based on the `simple_governance` module. However, we have established in section [Keepers for our app](module-keeper.md) that our module needs access to two other modules: the `bank` module and the `stake` module. We also need the `auth` module for basic account functionalities. Finally, we need access to the main multistore to declare the stores of each of the module we use.

## CLI and Rest server

We will need to add the newly created commands to our application. To do so, go to the `cmd` folder inside your root  directory:

```bash
// At root level of directory
cd cmd
```
`simplegovd` is the folder that stores the command for running the server daemon, whereas `simplegovcli` defines the commands of your application.

### Application CLI

**File: [`cmd/simplegovcli/maing.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/cmd/simplegovcli/main.go)**

To interact with our application, let us add the commands from the `simple_governance` module to our `simpleGov` application, as well as the pre-built SDK commands:

```go
//  cmd/simplegovcli/main.go
...
    rootCmd.AddCommand(
        client.GetCommands(
            simplegovcmd.GetCmdQueryProposal("proposals", cdc),
            simplegovcmd.GetCmdQueryProposals("proposals", cdc),
            simplegovcmd.GetCmdQueryProposalVotes("proposals", cdc),
            simplegovcmd.GetCmdQueryProposalVote("proposals", cdc),
        )...)
    rootCmd.AddCommand(
        client.PostCommands(
            simplegovcmd.PostCmdPropose(cdc),
            simplegovcmd.PostCmdVote(cdc),
        )...)
...
```

### Rest server

**File: [`cmd/simplegovd/main.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/cmd/simplegovd/main.go)**

The `simplegovd` command will run the daemon server as a background process. First, let us create some `utils` functions:

```go
//  cmd/simplegovd/main.go

// SimpleGovAppGenState sets up the app_state and appends the simpleGov app state
func SimpleGovAppGenState(cdc *codec.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {
    appState, err = server.SimpleAppGenState(cdc, appGenTxs)
    if err != nil {
        return
    }
    return
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
    return app.NewSimpleGovApp(logger, db)
}

func exportAppState(logger log.Logger, db dbm.DB) (json.RawMessage, error) {
    dapp := app.NewSimpleGovApp(logger, db)
    return dapp.ExportAppStateJSON()
}
```

Now, let us define the command for the daemon server within the `main()` function:

```go
//  cmd/simplegovd/main.go
func main() {
    cdc := app.MakeCodec()
    ctx := server.NewDefaultContext()

    rootCmd := &cobra.Command{
        Use:               "simplegovd",
        Short:             "Simple Governance Daemon (server)",
        PersistentPreRunE: server.PersistentPreRunEFn(ctx),
    }

    server.AddCommands(ctx, cdc, rootCmd,
        server.ConstructAppCreator(newApp, "simplegov"),
        server.ConstructAppExporter(exportAppState, "simplegov"))

    // prepare and add flags
    rootDir := os.ExpandEnv("$HOME/.simplegovd")
    executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
    executor.Execute()
}
```

## Makefile

The [Makefile](https://en.wikipedia.org/wiki/Makefile) compiles the Go program by defining a set of rules with targets and recipes. We'll need to add our application commands to it:

```
// Makefile
build_examples:
ifeq ($(OS),Windows_NT)
    ...
    go build $(BUILD_FLAGS) -o build/simplegovd.exe ./examples/simpleGov/cmd/simplegovd
    go build $(BUILD_FLAGS) -o build/simplegovcli.exe ./examples/simpleGov/cmd/simplegovcli
else
    ...
    go build $(BUILD_FLAGS) -o build/simplegovd ./examples/simpleGov/cmd/simplegovd
    go build $(BUILD_FLAGS) -o build/simplegovcli ./examples/simpleGov/cmd/simplegovcli
endif
...
install_examples:
    ...
    go install $(BUILD_FLAGS) ./examples/simpleGov/cmd/simplegovd
    go install $(BUILD_FLAGS) ./examples/simpleGov/cmd/simplegovcli
```

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
        capKeyMainStore:      sdk.NewKVStoreKey(bam.MainStoreKey),
        capKeyAccountStore:   sdk.NewKVStoreKey(auth.StoreKey),
        capKeyStakingStore:   sdk.NewKVStoreKey(stake.StoreKey),
        capKeySimpleGovStore: sdk.NewKVStoreKey("simpleGov"),
    }
```

- Instantiate the keepers. Note that keepers generally need access to other module's keepers. In this case, make sure you only pass an instance of the keeper for the functionality that is needed. If a keeper only needs to read in another module's store, a read-only keeper should be passed to it.

```go
app.bankKeeper = bank.NewBaseKeeper(app.accountKeeper)
app.stakeKeeper = simplestake.NewKeeper(app.capKeyStakingStore, app.bankKeeper,app.RegisterCodespace(simplestake.DefaultCodespace))
app.simpleGovKeeper = simpleGov.NewKeeper(app.capKeySimpleGovStore, app.bankKeeper, app.stakeKeeper, app.RegisterCodespace(simpleGov.DefaultCodespace))
```

- Declare the handlers.

```go
app.Router().
        AddRoute("bank", bank.NewHandler(app.bankKeeper)).
        AddRoute("simplestake", simplestake.NewHandler(app.stakeKeeper)).
        AddRoute("simpleGov", simpleGov.NewHandler(app.simpleGovKeeper))
```

- Initialize the application.

```go
// Initialize BaseApp.
    app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore, app.capKeySimpleGovStore, app.capKeyStakingStore)
    app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper))
    err := app.LoadLatestVersion(app.capKeyMainStore)
    if err != nil {
        cmn.Exit(err.Error())
    }
    return app
```

## Application codec

**File: [`app/app.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/app/app.go)**

Finally, we need to define the `MakeCodec()` function and register the concrete types and interface from the various modules.

```go
func MakeCodec() *codec.Codec {
    var cdc = codec.New()
    codec.RegisterCrypto(cdc) // Register crypto.
    sdk.RegisterCodec(cdc)    // Register Msgs
    bank.RegisterCodec(cdc)
    simplestake.RegisterCodec(cdc)
    simpleGov.RegisterCodec(cdc)

    // Register AppAccount
    cdc.RegisterInterface((*auth.Account)(nil), nil)
    cdc.RegisterConcrete(&types.AppAccount{}, "simpleGov/Account", nil)
    return cdc
}
```
