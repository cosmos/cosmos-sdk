##### Rest server

**File: [`cmd/simplegovd/main.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/cmd/simplegovd/main.go)**

The `simplegovd` command will run the daemon server as a background process. First, let us create some `utils` functions:

```go
//  cmd/simplegovd/main.go
// SimpleGovAppInit initial parameters
var SimpleGovAppInit = server.AppInit{
	AppGenState: SimpleGovAppGenState,
	AppGenTx:    server.SimpleAppGenTx,
}

// SimpleGovAppGenState sets up the app_state and appends the simpleGov app state
func SimpleGovAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {
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

	server.AddCommands(ctx, cdc, rootCmd, SimpleGovAppInit,
		server.ConstructAppCreator(newApp, "simplegov"),
		server.ConstructAppExporter(exportAppState, "simplegov"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.simplegovd")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	executor.Execute()
}
```