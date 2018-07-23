# Basecoin

As we've seen, the SDK provides a flexible yet comprehensive framework for building state
machines and defining their transitions, including authenticating transactions,
executing messages, controlling access to stores, and updating the validator set.

Until now, we have focused on building only isolated ABCI applications to
demonstrate and explain the various features and flexibilities of the SDK.
Here, we'll connect our ABCI application to Tendermint so we can run a full
blockchain node, and introduce command line and HTTP interfaces for interacting with it.

But first, let's talk about how source code should be laid out.

## Directory Structure

TODO

## Tendermint Node

Since the Cosmos-SDK is written in Go, Cosmos-SDK applications can be compiled
with Tendermint into a single binary. Of course, like any ABCI application, they
can also run as separate processes that communicate with Tendermint via socket.

For more details on what's involved in starting a Tendermint full node, see the
[NewNode](https://godoc.org/github.com/tendermint/tendermint/node#NewNode)
function in `github.com/tendermint/tendermint/node`.

The `server` package in the Cosmos-SDK simplifies
connecting an application with a Tendermint node.
For instance, the following `main.go` file will give us a complete full node
using the Basecoin application we built:

```go
//TODO imports

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "basecoind",
		Short:             "Basecoin Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, server.DefaultAppInit,
		server.ConstructAppCreator(newApp, "basecoin"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.basecoind")
	executor := cli.PrepareBaseCmd(rootCmd, "BC", rootDir)
	executor.Execute()
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewBasecoinApp(logger, db)
}
```

Note we utilize the popular [cobra library](https://github.com/spf13/cobra)
for the CLI, in concert with the [viper library](https://github.com/spf13/library)
for managing configuration. See our [cli library](https://github.com/tendermint/blob/master/tmlibs/cli/setup.go)
for more details.

TODO: compile and run the binary

Options for running the `basecoind` binary are effectively the same as for `tendermint`.
See [Using Tendermint](TODO) for more details.

## Clients

TODO
