# Upgrade Reference

This document provides a quick reference for the upgrades from `v0.50.x` to `v0.53.x` of Cosmos SDK.

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
          panic(fmt.Errorf("failed to register snapshot extension: %w", err))
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

<<<<<<< HEAD
  ```go
  func (app *App) Close() error {
      // ...
=======
:::warning

Using `protocolpool` will cause the following `x/distribution` handlers to return an error:


**QueryService**

- `CommunityPool`

**MsgService**

- `CommunityPoolSpend`
- `FundCommunityPool`

If you have services that rely on this functionality from `x/distribution`, please update them to use the `x/protocolpool` equivalents.

:::

⚠️Adding this module requires a `StoreUpgrade`⚠️
>>>>>>> 43f8fa113 (docs: explicit warnings about using external community pool with x/distribution (#24398))

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
