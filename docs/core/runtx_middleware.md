<!--
order: 12
-->

# RunTx recovery middleware

`BaseApp.runTx()` function handles Golang panics that might occur during transactions execution, for example, keeper has faced an invalid state and paniced.
Depending on the panic type different handler is used, for instance the default one prints an error log message.
Recovery middleware is used to add custom panic recovery for SDK application developers.

More context could be found in the corresponding [ADR-022](../architecture/adr-022-custom-panic-handling.md).

Implementation could be found in the [recovery.go](../../baseapp/recovery.go) file.

## Interface

```go
type RecoveryHandler func(recoveryObj interface{}) error
```

`recoveryObj` is a return value for `recover()` function from the `buildin` Golang package.

**Contract:**

- RecoveryHandler returns `nil` if `recoveryObj` wasn't handled and should be passed to the next recovery middleware;
- RecoveryHandler returns a non-nil `error` if `recoveryObj` was handled;

## Custom RecoveryHandler register

`BaseApp.AddRunTxRecoveryHandler(handlers ...RecoveryHandler)`

BaseApp method adds recovery middleware to the default recovery chain.

## Example

Lets assume we want to emit the "Consensus failure" chain state if some particular error occurred.

We have a module keeper that panics:

```go
func (k FooKeeper) Do(obj interface{}) {
    if obj == nil {
        // that shouldn't happen, we need to crash the app
        err := sdkErrors.Wrap(fooTypes.InternalError, "obj is nil")
        panic(err)
    }
}
```

By default that panic would be recovered and an error message will be printed to log. To override that behaviour we should register a custom RecoveryHandler:

```go
// SDK application constructor
customHandler := func(recoveryObj interface{}) error {
    err, ok := recoveryObj.(error)
    if !ok {
        return nil
    }

    if fooTypes.InternalError.Is(err) {
        panic(fmt.Errorf("FooKeeper did panic with error: %w", err))
    }

    return nil
}

baseApp := baseapp.NewBaseApp(...)
baseApp.AddRunTxRecoveryHandler(customHandler)
```

## Next {hide}

Learn about the [IBC](./../ibc/README.md) protocol {hide}
