# BaseApp

The BaseApp defines the foundational implementation for a basic ABCI application
so that your Cosmos-SDK application can communicate with an underlying
Tendermint node.

The BaseApp is composed of many internal components. Some of the most important
include the `CommitMultiStore` and its internal state. The internal state is
essentially two sub-states, both of which are used for transaction execution
during different phases, `CheckTx` and `DeliverTx` respectively. During block
commitment, only the `DeliverTx` is persisted.

## Transaction Life Cycle

During the execution of a transaction, it may pass through both `CheckTx` and
`DeliverTx` as defined in the ABCI specification. `CheckTx` is executed by the
proposing validator and is used for the Tendermint mempool. Notably, both
`CheckTx` and `DeliverTx` execute the application's AnteHandler (if defined),
where the AnteHandler is responsible for pre-message validation checks such as
account and signature validation as well as fee deduction and collection. It is
important to note that the AnteHandler does, and is expected to, perform various
state transitions such as nonce incrimination and fee deduction.

### CheckTx

During the execution of `CheckTx`, first the AnteHandler is executed. If the
AnteHandler fails or panics then the transaction fails. After the successful
execution of theAnteHandler, the messages are executed. If all the messages
successfully executed, then the state transitions as a result of executing the
messages are persisted.

It is important to note that the state transitions due to the AnteHandler are
persisted between subsequent calls of `CheckTx`. Also, currently the AnteHandler
also handles making sure the sender has included enough fees as these mechanics
differ per validator, it must be checked in `CheckTx` as to not cause conflicting
state transitions during consensus.

### DeliverTx

The transaction execution during `DeliverTx` operates in a similar fashion to
`CheckTx`. However, state transitions caused by the AnteHandler are persisted
unless the transaction fails.

It is possible that a malicious proposer may send a transaction that would fail
`CheckTx` but send it anyway causing other full nodes to execute it during
`DeliverTx`. Because of this, we do not want to brake the invariant that state
transitions cannot occur for transactions that would otherwise fail `CheckTx`,
hence we do not persist state in such cases.
