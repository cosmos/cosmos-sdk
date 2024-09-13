# server v2 integration tests

## audit usages of [NewIntegrationApp](../../testutil/integration/router.go#L46)


* x/*
    * create and register query and message servers 
    * call `App.QueryHelper` in setup, and may call in test
    * make use of `sdk.Context`. a `context.Context` could be substituted except where otherwise noted
* [x/auth](./auth/keeper/msg_server_test.go#L122) 
    * calls `App.RunMsg` in test
    * calls `sdk.Context` `GasMeter()` through [testdata.DeterministicIterations](../../testutil/testdata/grpc_query.go#L73)
* [x/bank](./bank/keeper/deterministic_test.go#L122)
    * calls `sdk.Context` `GasMeter()` through [testdata.DeterministicIterations](../../testutil/testdata/grpc_query.go#L73)
* [x/distribution](./distribution/keeper/msg_server_test.go#L170)
    * calls `App.RunMsg` in test
    * calls `BaseApp.LastBlockHeight()`. delegates to store/v1 meta info store.
    * calls `sdk.Context.CometInfo()`. can be replaced with `CometInfoService`
* [x/evidence](./evidence//keeper/infraction_test.go#L164)
    * calls `BaseApp.StoreConsenusParams`, `BaseApp.GetConsensusParams` in test.
    * mutates `sdk.Context` with `WithIsCheckTx`, `WithBlockHeight`, `WithHeaderInfo`
    * calls `sdk.Context` `GetBlockHeight()`
* [x/gov](./gov/keeper/keeper_test.go#L150)
* x/slashing
    * calls `App.RunMsg` in test
    * calls `sdk.Context` `BlockHeight()`
    * mutates `sdk.Context` with `WithBlockHeight`, `WithHeaderInfo`
* x/staking
    * calls `sdk.Context` `GasMeter()` through [testdata.DeterministicIterations](../../testutil/testdata/grpc_query.go#L73)

## poc notes

testutil/integration can (and should) be moved to tests/integration.

## differences from v1

- a v1 app calls commit only after cometbft has committed a block and calls the ABCI life cycle function Commit() 
- a v2 app calls commit within FinalizeBlock.

## simulations

In addition to `NewIntegrationApp` many modules also make use of [testutil/sims.SetupWithConfiguration](../../testutil/sims/app_helpers.go#L145) to create a `runtime.App` for simulation testing. For example 
[createTestSuite in bank](bank/app_test.go#L81)
