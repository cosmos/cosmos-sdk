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
"main" (`baseApp.MainStoreKey`) - it holds the latest block header, from which we can find and load the
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
proposing validator and is used for the Tendermint mempool for all full nodes.

Both `CheckTx` and `DeliverTx` execute the application's AnteHandler (if
defined), where the AnteHandler is responsible for pre-message validation
checks such as account and signature validation, fee deduction and collection,
and incrementing sequence numbers.

### CheckTx

During the execution of `CheckTx`, only the AnteHandler is executed.

State transitions due to the AnteHandler are persisted between subsequent calls
of `CheckTx` in the check-tx state, unless the AnteHandler fails and aborts.

### DeliverTx

During the execution of `DeliverTx`, the AnteHandler and Handler is executed.

The transaction execution during `DeliverTx` operates in a similar fashion to
`CheckTx`. However, state transitions that occur during the AnteHandler are
persisted even when the following Handler processing logic fails.

It is possible that a malicious proposer may include a transaction in a block
that fails the AnteHandler.  In this case, all state transitions for the
offending transaction are discarded.


## Other ABCI Messages

Besides `CheckTx` and `DeliverTx`, BaseApp handles the following ABCI messages.

### Info
TODO complete description

### SetOption
TODO complete description

### Query
TODO complete description

### InitChain
TODO complete description

During chain initialization InitChain runs the initialization logic directly on
the CommitMultiStore. The deliver and check states are initialized with the
ChainID.

Note that we do not commit after InitChain, so BeginBlock for block 1 starts
from the deliver state as initialized by InitChain.

### BeginBlock
TODO complete description

### EndBlock
TODO complete description

### Commit
TODO complete description


## Gas Management

### Gas: InitChain

During InitChain, the block gas meter is initialized with an infinite amount of
gas to run any genesis transactions.

Additionally, the InitChain request message includes ConsensusParams as
declared in the genesis.json file.

### Gas: BeginBlock

The block gas meter is reset during BeginBlock for the deliver state.  If no
maximum block gas is set within baseapp then an infinite gas meter is set,
otherwise a gas meter with `ConsensusParam.BlockSize.MaxGas` is initialized.

### Gas: DeliverTx

Before the transaction logic is run, the `BlockGasMeter` is first checked to
see if any gas remains. If no gas remains, then `DeliverTx` immediately returns
an error.

After the transaction has been processed, the used gas (up to the transaction
gas limit) is deducted from the BlockGasMeter. If the remaining gas exceeds the
meter's limits, then DeliverTx returns an error and the transaction is not
committed.
