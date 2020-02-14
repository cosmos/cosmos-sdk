<!--
order: 3
synopsis: "`Message`s and `Queries` are the two primary objects handled by modules. Most of the core components defined in a module, like `handler`s, `keeper`s and `querier`s, exist to process `message`s and `queries`."
-->

# Messages and Queries

## Pre-requisite Readings {hide}

- [Introduction to SDK Modules](./intro.md) {prereq}

## Messages

`Message`s are objects whose end-goal is to trigger state-transitions. They are wrapped in [transactions](../core/transactions.md), which may contain one or multiple of them. 

When a transaction is relayed from the underlying consensus engine to the SDK application, it is first decoded by [`baseapp`](../core/baseapp.md). Then, each `message` contained in the transaction is extracted and routed to the appropriate module via `baseapp`'s `router` so that it can be processed by the module's [`handler`](./handler.md). For a more detailed explanation of the lifecycle of a transaction, click [here](../basics/tx-lifecycle.md). 

Defining `message`s is the responsibility of module developers. Typically, they are defined in a `./internal/types/msgs.go` file inside the module's folder. The `message`'s type definition usually includes a list of parameters needed to process the message that will be provided by end-users when they want to create a new transaction containing said `message`.

```go
// Example of a message type definition

type MsgSubmitProposal struct {
	Content        Content        `json:"content" yaml:"content"`
	InitialDeposit sdk.Coins      `json:"initial_deposit" yaml:"initial_deposit"` 
	Proposer       sdk.AccAddress `json:"proposer" yaml:"proposer"`               
}
```

The `Msg` is typically accompanied by a standard constructor function, that is called from one of the [module's interface](./module-interfaces.md). `message`s also need to implement the [`Msg`] interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/tx_msg.go#L7-L29

It contains the following methods:

- `Route() string`: Name of the route for this message. Typically all `message`s in a module have the same route, which is most often the module's name.
- `Type() string`: Type of the message, used primarly in [events](../core/events.md). This should return a message-specific `string`, typically the denomination of the message itself.
- `ValidateBasic() Error`: This method is called by `baseapp` very early in the processing of the `message` (in both [`CheckTx`](../core/baseapp.md#checktx) and [`DeliverTx`](../core/baseapp.md#delivertx)), in order to discard obviously invalid messages. `ValidateBasic` should only include *stateless* checks, i.e. checks that do not require access to the state. This usually consists in checking that the message's parameters are correctly formatted and valid (i.e. that the `amount` is strictly positive for a transfer).
- `GetSignBytes() []byte`: Return the canonical byte representation of the message. Used to generate a signature. 
- `GetSigners() []AccAddress`: Return the list of signers. The SDK will make sure that each `message` contained in a transaction is signed by all the signers listed in the list returned by this method. 

See an example implementation of a `message` from the `nameservice` module:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/internal/types/msgs.go#L10-L51

## Queries

A `query` is a request for information made by end-users of applications through an interface and processed by a full-node. A `query` is received by a full-node through its consensus engine and relayed to the application via the ABCI. It is then routed to the appropriate module via `baseapp`'s `queryrouter` so that it can be processed by the module's [`querier`](./querier.md). For a deeper look at the lifecycle of a `query`, click [here](../interfaces/query-lifecycle.md). 

Contrary to `message`s, there is usually no specific `query` object defined by module developers. Instead, the SDK takes the simpler approach of using a simple `path` to define each `query`. The `path` contains the `query` type and all the arguments needed in order to process it. For most module queries, the `path` should look like the following:

```
queryCategory/queryRoute/queryType/arg1/arg2/...
```

where:

- `queryCategory` is the category of the `query`, typically `custom` for module queries. It is used to differentiate between different kinds of queries within `baseapp`'s [`Query` method](../core/baseapp.md#query).
- `queryRoute` is used by `baseapp`'s [`queryRouter`](../core/baseapp.md#query-routing) to map the `query` to its module. Usually, `queryRoute` should be the name of the module.
- `queryType` is used by the module's [`querier`](./querier.md) to map the `query` to the appropriate `querier function` within the module. 
- `args` are the actual arguments needed to process the `query`. They are filled out by the end-user. Note that for bigger queries, you might prefer passing arguments in the `Data` field of the request `req` instead of the `path`. 

The `path` for each `query` must be defined by the module developer in the module's [command-line interface file](./module-interfaces.md#query-commands).Overall, there are 3 mains components module developers need to implement in order to make the subset of the state defined by their module queryable:

- A [`querier`](./querier.md), to process the `query` once it has been [routed to the module](../core/baseapp.md#query-routing). 
- [Query commands](./module-interfaces.md#query-commands) in the module's CLI file, where the `path` for each `query` is specified. 
- `query` return types. Typically defined in a file `internal/types/querier.go`, they specify the result type of each of the module's `queries`. These custom types must implement the `String()` method of [`fmt.Stringer`](https://golang.org/pkg/fmt/#Stringer). 

See an example of `query` return types from the `nameservice` module:

+++ https://github.com/cosmos/sdk-tutorials/blob/c6754a1e313eb1ed973c5c91dcc606f2fd288811/x/nameservice/internal/types/querier.go#L5-L21

## Next {hide}

Learn about [`handler`s](./handler.md) {hide}