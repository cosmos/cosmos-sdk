# Upgrading Cosmos SDK

### BaseApp

#### Nested Messages Simulation

Now it is possible to simulate the nested messages of a message, providing developers with a powerful tool for
testing and predicting the behavior of complex transactions. This feature allows for a more comprehensive
evaluation of gas consumption, state changes, and potential errors that may occur when executing nested
messages. However, it's important to note that while the simulation can provide valuable insights, it does not
guarantee the correct execution of the nested messages in the future. Factors such as changes in the
blockchain state or updates to the protocol could potentially affect the actual execution of these nested
messages when the transaction is finally processed on the network.

For example, consider a governance proposal that includes nested messages to update multiple protocol
parameters. At the time of simulation, the blockchain state may be suitable for executing all these nested
messages successfully. However, by the time the actual governance proposal is executed (which could be days or
weeks later), the blockchain state might have changed significantly. As a result, while the simulation showed
a successful execution, the actual governance proposal might fail when it's finally processed.

By default, when simulating transactions, the gas cost of nested messages is not calculated. This means that
only the gas cost of the top-level message is considered. However, this behavior can be customized using the
`SetIncludeNestedMsgsGas` option when building the BaseApp. By providing a list of message types to this option,
you can specify which messages should have their nested message gas costs included in the simulation. This
allows for more accurate gas estimation for transactions involving specific message types that contain nested
messages, while maintaining the default behavior for other message types.

Here is an example on how `SetIncludeNestedMsgsGas` option could be set to calculate the gas of a gov proposal
nested messages:
```go
baseAppOptions = append(baseAppOptions, baseapp.SetIncludeNestedMsgsGas([]sdk.Message{&gov.MsgSubmitProposal{}}))
// ...
app.App = appBuilder.Build(db, traceStore, baseAppOptions...)
```

This guide provides instructions for upgrading to specific versions of Cosmos SDK.
Note, always read the **SimApp** section for more information on application wiring updates.

## [v0.53.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.0)

#### Unordered Transactions

The Cosmos SDK now supports unordered transactions. This means that transactions
can be executed in any order and doesn't require the client to deal with or manage
nonces. This also means the order of execution is not guaranteed.

Unordered transactions are automatically enabled when using `depinject` / app di, simply supply the `servertypes.AppOptions` in `app.go`:

```diff
	depinject.Supply(
+		// supply the application options
+		appOpts,
		// supply the logger
		logger,
	)
```

<details>
<summary>Step-by-step Wiring </summary>
If you are still using the legacy wiring, you must enable unordered transactions manually:

* Update the `App` constructor to create, load, and save the unordered transaction
  manager.

  ```go
  func NewApp(...) *App {
      // ...

      // create, start, and load the unordered tx manager
      utxDataDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data")
      app.UnorderedTxManager = unorderedtx.NewManager(utxDataDir)
      app.UnorderedTxManager.Start()

      if err := app.UnorderedTxManager.OnInit(); err != nil {
          panic(fmt.Errorf("failed to initialize unordered tx manager: %w", err))
      }
  }
  ```

* Add the decorator to the existing AnteHandler chain, which should be as early
  as possible.

  ```go
  anteDecorators := []sdk.AnteDecorator{
      ante.NewSetUpContextDecorator(),
      // ...
      ante.NewUnorderedTxDecorator(unorderedtx.DefaultMaxTimeoutDuration, options.TxManager, options.Environment),
      // ...
  }

  return sdk.ChainAnteDecorators(anteDecorators...), nil
  ```

* If the App has a SnapshotManager defined, you must also register the extension
  for the TxManager.

  ```go
  if manager := app.SnapshotManager(); manager != nil {
      err := manager.RegisterExtensions(unorderedtx.NewSnapshotter(app.UnorderedTxManager))
      if err != nil {
          panic(fmt.Errorf("failed to register snapshot extension: %s", err))
      }
  }
  ```

* Create or update the App's `Preblocker()` method to call the unordered tx
  manager's `OnNewBlock()` method.

  ```go
  ...
  app.SetPreblocker(app.PreBlocker)
  ...

  func (app *SimApp) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
      app.UnorderedTxManager.OnNewBlock(ctx.BlockTime())
      return app.ModuleManager.PreBlock(ctx, req)
  }
  ```

* Create or update the App's `Close()` method to close the unordered tx manager.
  Note, this is critical as it ensures the manager's state is written to file
  such that when the node restarts, it can recover the state to provide replay
  protection.

  ```go
  func (app *App) Close() error {
      // ...

      // close the unordered tx manager
      if e := app.UnorderedTxManager.Close(); e != nil {
          err = errors.Join(err, e)
      }

      return err
  }
  ```

</details>

To submit an unordered transaction, the client must set the `unordered` flag to
`true` and ensure a reasonable `timeout_height` is set. The `timeout_height` is
used as a TTL for the transaction and is used to provide replay protection. See
[ADR-070](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-070-unordered-transactions.md) for more details.
