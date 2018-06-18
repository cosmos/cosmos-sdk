# BaseApp

The BaseApp is an abstraction over the [Tendermint
ABCI](https://github.com/tendermint/abci) that
simplifies application development by handling common low-level concerns.
It serves as the mediator between the two key components of an SDK app: the store
and the message handlers.

The BaseApp implements the
[`abci.Application`](https://godoc.org/github.com/tendermint/abci/types#Application) interface. 
It uses a `MultiStore` to manage the state, a `Router` for transaction handling, and 
`Set` methods to specify functions to run at the beginning and end of every
block. 

Every SDK app begins with a BaseApp:

```
app := baseapp.NewBaseApp(appName, cdc, logger, db),
```
