# Messages and Transactions

## Messages
**Messages** describe possible actions within a module. Note: these messages are not to be confused with [ABCI Messages](https://tendermint.com/docs/spec/abci/abci.html#messages) which define interactions between Tendermint and the application.

Developers define the specific messages for each application module by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface.

### ValidateBasic
### Route 


## Transactions
**[Transactions](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L36-L43)** are comprised of one or multiple `Msg`s and trigger state changes.

Every module has a `tx.go` file with a `GetTxCmd` function that returns the transaction commands for that module.
