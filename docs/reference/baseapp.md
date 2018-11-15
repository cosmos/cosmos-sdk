# BaseApp

The BaseApp defines the foundational implementation for a basic ABCI application
so that your Cosmos-SDK application can communicate with an underlying
Tendermint node.

The BaseApp is composed of many internal components. Some of the most important
include the `CommitMultiStore` and its internal state. The internal state is
essentially two sub-states, both of which are used for transaction execution
during different phases, `CheckTx` and `DeliverTx` respectively. During block
commitment, only the `DeliverTx` is persisted.

The BaseApp requires stores to be mounted via capabilities keys - handlers can
only access stores they're given the key to. The `baseApp` ensures all stores are
properly loaded, cached, and committed. One mounted store is considered the
"main" - it holds the latest block header, from which we can find and load the
most recent state.

The BaseApp distinguishes between two handler types - the `AnteHandler` and the
`MsgHandler`. The former is a global validity check (checking nonces, sigs and
sufficient balances to pay fees, e.g. things that apply to all transaction from
all modules), the later is the full state transition function.

During `CheckTx` the state transition function is only applied to the `checkTxState`
and should return before any expensive state transitions are run
(this is up to each developer). It also needs to return the estimated gas cost.

During `DeliverTx` the state transition function is applied to the blockchain
state and the transactions need to be fully executed.

The BaseApp is responsible for managing the context passed into handlers -
it makes the block header available and provides the right stores for `CheckTx`
and `DeliverTx`. BaseApp is completely agnostic to serialization formats.

## Transaction Life Cycle

During the execution of a transaction, it may pass through both `CheckTx` and
`DeliverTx` as defined in the ABCI specification. `CheckTx` is executed by the
proposing validator and is used for the Tendermint mempool. Notably, both
`CheckTx` and `DeliverTx` execute the application's AnteHandler (if defined),
where the AnteHandler is responsible for pre-message validation checks such as
account and signature validation as well as fee deduction and collection. It is
important to note that the AnteHandler does, and is expected to, perform various
state transitions such as incrementing nonces and deducting fees.

### CheckTx

During the execution of `CheckTx`, only the AnteHandler is executed. If the
AnteHandler fails or panics then the transaction fails.

It is important to note that the state transitions due to the AnteHandler are
persisted between subsequent calls of `CheckTx`. Also, currently the AnteHandler
also handles making sure the sender has included enough fees as validators
(and non-validator nodes) can set their own gas-price for validating transactions.
This must be checked in `CheckTx` as to not cause conflicting state transitions
during consensus.

### DeliverTx

The transaction execution during `DeliverTx` operates in a similar fashion to
`CheckTx`. However, state transitions caused by the AnteHandler are persisted
unless the transaction fails in addition to the messages being executed. 

It is possible that a malicious proposer may send a transaction that would fail
`CheckTx` but send it anyway causing other full nodes to execute it during
`DeliverTx`.

Because of this, we do not want to brake the invariant that state transitions
cannot occur for transactions that would otherwise fail `CheckTx`, hence we do
not persist state in such cases.
