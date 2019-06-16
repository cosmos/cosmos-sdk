# Messages and Transactions

## Messages
**Messages** describe possible actions within a module. Note: these messages are not to be confused with [ABCI Messages](https://tendermint.com/docs/spec/abci/abci.html#messages) which define interactions between Tendermint and the application.

Developers define the specific messages for each application module by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface. They also define [`Handler`](https://github.com/cosmos/cosmos-sdk/blob/1cfc868d86a152b523443154c8723de832dbec81/types/handler.go#L4)s that execute the actions for each message and return the [`Result`](https://github.com/cosmos/cosmos-sdk/blob/1cfc868d86a152b523443154c8723de832dbec81/types/result.go#L14-L37). An [`AnteHandler`](https://github.com/cosmos/cosmos-sdk/blob/1cfc868d86a152b523443154c8723de832dbec81/types/handler.go#L8) can also be defined to execute a message's actions in simulation mode (i.e. without persisting state changes) to perform checks.


## Transactions
**[Transactions](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L36-L43)** are comprised of one or multiple `Msg`s and trigger state changes.

Every module has a `tx.go` file with a `GetTxCmd` function that returns the transaction commands for that module.
