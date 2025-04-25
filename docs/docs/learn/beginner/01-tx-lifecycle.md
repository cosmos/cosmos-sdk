---
sidebar_position: 1
---

# Transaction Lifecycle

:::note Synopsis
This document describes the lifecycle of a transaction from creation to committed state changes. Transaction definition is described in a [different doc](../advanced/01-transactions.md). The transaction is referred to as `Tx`.
:::

:::note Pre-requisite Readings

* [Anatomy of a Cosmos SDK Application](./00-app-anatomy.md)
:::

## Creation

### Transaction Creation

One of the main application interfaces is the command-line interface. The transaction `Tx` can be created by the user inputting a command in the following format from the [command-line](../advanced/07-cli.md), providing the type of transaction in `[command]`, arguments in `[args]`, and configurations such as gas prices in `[flags]`:

```bash
[appname] tx [command] [args] [flags]
```

This command automatically **creates** the transaction, **signs** it using the account's private key, and **broadcasts** it to the specified peer node.

There are several required and optional flags for transaction creation. The `--from` flag specifies which [account](./03-accounts.md) the transaction is originating from. For example, if the transaction is sending coins, the funds are drawn from the specified `from` address.

#### Gas and Fees

Additionally, there are several [flags](../advanced/07-cli.md) users can use to indicate how much they are willing to pay in [fees](./04-gas-fees.md):

* `--gas` refers to how much [gas](./04-gas-fees.md), which represents computational resources, `Tx` consumes. Gas is dependent on the transaction and is not precisely calculated until execution, but can be estimated by providing `auto` as the value for `--gas`.
* `--gas-adjustment` (optional) can be used to scale `gas` up in order to avoid underestimating. For example, users can specify their gas adjustment as 1.5 to use 1.5 times the estimated gas.
* `--gas-prices` specifies how much the user is willing to pay per unit of gas, which can be one or multiple denominations of tokens. For example, `--gas-prices=0.025uatom, 0.025upho` means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
* `--fees` specifies how much in fees the user is willing to pay in total.
* `--timeout-height` specifies a block timeout height to prevent the tx from being committed past a certain height.

The ultimate value of the fees paid is equal to the gas multiplied by the gas prices. In other words, `fees = ceil(gas * gasPrices)`. Thus, since fees can be calculated using gas prices and vice versa, the users specify only one of the two.

Later, validators decide whether to include the transaction in their block by comparing the given or calculated `gas-prices` to their local `min-gas-prices`. `Tx` is rejected if its `gas-prices` is not high enough, so users are incentivized to pay more.

#### Unordered Transactions

With Cosmos SDK v0.53.0, users may send unordered transactions to chains that have this feature enabled.
The following flags allow a user to build an unordered transaction from the CLI.

* `--unordered` specifies that this transaction should be unordered.
* `--timeout-duration` specifies the amount of time the unordered transaction should be valid in the mempool. The transaction's unordered nonce will be set to the time of transaction creation + timeout duration.

#### CLI Example

Users of the application `app` can enter the following command into their CLI to generate a transaction to send 1000uatom from a `senderAddress` to a `recipientAddress`. The command specifies how much gas they are willing to pay: an automatic estimate scaled up by 1.5 times, with a gas price of 0.025uatom per unit gas.

```bash
appd tx send <recipientAddress> 1000uatom --from <senderAddress> --gas auto --gas-adjustment 1.5 --gas-prices 0.025uatom
```

#### Other Transaction Creation Methods

The command-line is an easy way to interact with an application, but `Tx` can also be created using a [gRPC or REST interface](../advanced/06-grpc_rest.md) or some other entry point defined by the application developer. From the user's perspective, the interaction depends on the web interface or wallet they are using (e.g. creating `Tx` using [Lunie.io](https://lunie.io/#/) and signing it with a Ledger Nano S).

## Addition to Mempool

Each full-node (running CometBFT) that receives a `Tx` sends an [ABCI message](https://docs.cometbft.com/v0.37/spec/p2p/messages/),
`CheckTx`, to the application layer to check for validity, and receives an `abci.ResponseCheckTx`. If the `Tx` passes the checks, it is held in the node's
[**Mempool**](https://docs.cometbft.com/v0.37/spec/p2p/messages/mempool/), an in-memory pool of transactions unique to each node, pending inclusion in a block - honest nodes discard a `Tx` if it is found to be invalid. Prior to consensus, nodes continuously check incoming transactions and gossip them to their peers.

### Types of Checks

The full-nodes perform stateless, then stateful checks on `Tx` during `CheckTx`, with the goal to
identify and reject an invalid transaction as early on as possible to avoid wasted computation.

**_Stateless_** checks do not require nodes to access state - light clients or offline nodes can do
them - and are thus less computationally expensive. Stateless checks include making sure addresses
are not empty, enforcing nonnegative numbers, and other logic specified in the definitions.

**_Stateful_** checks validate transactions and messages based on a committed state. Examples
include checking that the relevant values exist and can be transacted with, the address
has sufficient funds, and the sender is authorized or has the correct ownership to transact.
At any given moment, full-nodes typically have [multiple versions](../advanced/00-baseapp.md#state-updates)
of the application's internal state for different purposes. For example, nodes execute state
changes while in the process of verifying transactions, but still need a copy of the last committed
state in order to answer queries - they should not respond using state with uncommitted changes.

In order to verify a `Tx`, full-nodes call `CheckTx`, which includes both _stateless_ and _stateful_
checks. Further validation happens later in the [`DeliverTx`](#delivertx) stage. `CheckTx` goes
through several steps, beginning with decoding `Tx`.

### Decoding

When `Tx` is received by the application from the underlying consensus engine (e.g. CometBFT ), it is still in its [encoded](../advanced/05-encoding.md) `[]byte` form and needs to be unmarshaled in order to be processed. Then, the [`runTx`](../advanced/00-baseapp.md#runtx-antehandler-runmsgs-posthandler) function is called to run in `runTxModeCheck` mode, meaning the function runs all checks but exits before executing messages and writing state changes.

### ValidateBasic (deprecated)

Messages ([`sdk.Msg`](../advanced/01-transactions.md#messages)) are extracted from transactions (`Tx`). The `ValidateBasic` method of the `sdk.Msg` interface implemented by the module developer is run for each transaction.
To discard obviously invalid messages, the `BaseApp` type calls the `ValidateBasic` method very early in the processing of the message in the [`CheckTx`](../advanced/00-baseapp.md#checktx) and [`DeliverTx`](../advanced/00-baseapp.md#delivertx) transactions.
`ValidateBasic` can include only **stateless** checks (the checks that do not require access to the state). 

:::warning
The `ValidateBasic` method on messages has been deprecated in favor of validating messages directly in their respective [`Msg` services](../../build/building-modules/03-msg-services.md#Validation).

Read [RFC 001](https://docs.cosmos.network/main/rfc/rfc-001-tx-validation) for more details.
:::

:::note
`BaseApp` still calls `ValidateBasic` on messages that implements that method for backwards compatibility.
:::

#### Guideline

`ValidateBasic` should not be used anymore. Message validation should be performed in the `Msg` service when [handling a message](../../build/building-modules/msg-services#Validation) in a module Msg Server.

### AnteHandler

`AnteHandler`s even though optional, are in practice very often used to perform signature verification, gas calculation, fee deduction, and other core operations related to blockchain transactions.

A copy of the cached context is provided to the `AnteHandler`, which performs limited checks specified for the transaction type. Using a copy allows the `AnteHandler` to do stateful checks for `Tx` without modifying the last committed state, and revert back to the original if the execution fails.

For example, the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/main/x/auth/spec) module `AnteHandler` checks and increments sequence numbers, checks signatures and account numbers, and deducts fees from the first signer of the transaction - all state changes are made using the `checkState`.

:::warning
Ante handlers only run on a transaction. If a transaction embed multiple messages (like some x/authz, x/gov transactions for instance), the ante handlers only have awareness of the outer message. Inner messages are mostly directly routed to the [message router](https://docs.cosmos.network/main/learn/advanced/baseapp#msg-service-router) and will skip the chain of ante handlers. Keep that in mind when designing your own ante handler.
:::

### Gas

The [`Context`](../advanced/02-context.md), which keeps a `GasMeter` that tracks how much gas is used during the execution of `Tx`, is initialized. The user-provided amount of gas for `Tx` is known as `GasWanted`. If `GasConsumed`, the amount of gas consumed during execution, ever exceeds `GasWanted`, the execution stops and the changes made to the cached copy of the state are not committed. Otherwise, `CheckTx` sets `GasUsed` equal to `GasConsumed` and returns it in the result. After calculating the gas and fee values, validator-nodes check that the user-specified `gas-prices` is greater than their locally defined `min-gas-prices`.

### Discard or Addition to Mempool

If at any point during `CheckTx` the `Tx` fails, it is discarded and the transaction lifecycle ends
there. Otherwise, if it passes `CheckTx` successfully, the default protocol is to relay it to peer
nodes and add it to the Mempool so that the `Tx` becomes a candidate to be included in the next block.

The **mempool** serves the purpose of keeping track of transactions seen by all full-nodes.
Full-nodes keep a **mempool cache** of the last `mempool.cache_size` transactions they have seen, as a first line of
defense to prevent replay attacks. Ideally, `mempool.cache_size` is large enough to encompass all
of the transactions in the full mempool. If the mempool cache is too small to keep track of all
the transactions, `CheckTx` is responsible for identifying and rejecting replayed transactions.

Currently existing preventative measures include fees and a `sequence` (nonce) counter to distinguish
replayed transactions from identical but valid ones. If an attacker tries to spam nodes with many
copies of a `Tx`, full-nodes keeping a mempool cache reject all identical copies instead of running
`CheckTx` on them. Even if the copies have incremented `sequence` numbers, attackers are
disincentivized by the need to pay fees.

Validator nodes keep a mempool to prevent replay attacks, just as full-nodes do, but also use it as
a pool of unconfirmed transactions in preparation of block inclusion. Note that even if a `Tx`
passes all checks at this stage, it is still possible to be found invalid later on, because
`CheckTx` does not fully validate the transaction (that is, it does not actually execute the messages).

## Inclusion in a Block

Consensus, the process through which validator nodes come to agreement on which transactions to
accept, happens in **rounds**. Each round begins with a proposer creating a block of the most
recent transactions and ends with **validators**, special full-nodes with voting power responsible
for consensus, agreeing to accept the block or go with a `nil` block instead. Validator nodes
execute the consensus algorithm, such as [CometBFT](https://docs.cometbft.com/v0.37/spec/consensus/),
confirming the transactions using ABCI requests to the application, in order to come to this agreement.

The first step of consensus is the **block proposal**. One proposer amongst the validators is chosen
by the consensus algorithm to create and propose a block - in order for a `Tx` to be included, it
must be in this proposer's mempool.

## State Changes

The next step of consensus is to execute the transactions to fully validate them. All full-nodes
that receive a block proposal from the correct proposer execute the transactions by calling the ABCI function `FinalizeBlock`. 
As mentioned throughout the documentation `BeginBlock`, `ExecuteTx` and `EndBlock` are called within FinalizeBlock. 
Although every full-node operates individually and locally, the outcome is always consistent and unequivocal. This is because the state changes brought about by the messages are predictable, and the transactions are specifically sequenced in the proposed block.

```text
        --------------------------
        | Receive Block Proposal |
        --------------------------
                    |
                    v
        -------------------------
        |     FinalizeBlock      |
        -------------------------
                    |
                    v
            -------------------
            |   BeginBlock     | 
            -------------------
                    |
                    v
            --------------------
            | ExecuteTx(tx0)   |
            | ExecuteTx(tx1)   |
            | ExecuteTx(tx2)   |
            | ExecuteTx(tx3)   |
            |       .          |
            |       .          |
            |       .          |
            -------------------
                    |
                    v
            --------------------
            |    EndBlock      |
            --------------------
                    |
                    v
        -------------------------
        |       Consensus        |
        -------------------------
                    |
                    v
        -------------------------
        |         Commit         |
        -------------------------
```

### Transaction Execution

The `FinalizeBlock` ABCI function defined in [`BaseApp`](../advanced/00-baseapp.md) does the bulk of the
state transitions: it is run for each transaction in the block in sequential order as committed
to during consensus. Under the hood, transaction execution is almost identical to `CheckTx` but calls the
[`runTx`](../advanced/00-baseapp.md#runtx) function in deliver mode instead of check mode.
Instead of using their `checkState`, full-nodes use `finalizeblock`:

* **Decoding:** Since `FinalizeBlock` is an ABCI call, `Tx` is received in the encoded `[]byte` form.
  Nodes first unmarshal the transaction, using the [`TxConfig`](./app-anatomy#register-codec) defined in the app, then call `runTx` in `execModeFinalize`, which is very similar to `CheckTx` but also executes and writes state changes.

* **Checks and `AnteHandler`:** Full-nodes call `validateBasicMsgs` and `AnteHandler` again. This second check
  happens because they may not have seen the same transactions during the addition to Mempool stage 
  and a malicious proposer may have included invalid ones. One difference here is that the
  `AnteHandler` does not compare `gas-prices` to the node's `min-gas-prices` since that value is local
  to each node - differing values across nodes yield nondeterministic results.

* **`MsgServiceRouter`:** After `CheckTx` exits, `FinalizeBlock` continues to run
  [`runMsgs`](../advanced/00-baseapp.md#runtx-antehandler-runmsgs-posthandler) to fully execute each `Msg` within the transaction.
  Since the transaction may have messages from different modules, `BaseApp` needs to know which module
  to find the appropriate handler. This is achieved using `BaseApp`'s `MsgServiceRouter` so that it can be processed by the module's Protobuf [`Msg` service](../../build/building-modules/03-msg-services.md).
  For `LegacyMsg` routing, the `Route` function is called via the [module manager](../../build/building-modules/01-module-manager.md) to retrieve the route name and find the legacy [`Handler`](../../build/building-modules/03-msg-services.md#handler-type) within the module.
  
* **`Msg` service:** Protobuf `Msg` service is responsible for executing each message in the `Tx` and causes state transitions to persist in `finalizeBlockState`.

* **PostHandlers:** [`PostHandler`](../advanced/00-baseapp.md#posthandler)s run after the execution of the message. If they fail, the state change of `runMsgs`, as well of `PostHandlers`, are both reverted.

* **Gas:** While a `Tx` is being delivered, a `GasMeter` is used to keep track of how much
  gas is being used; if execution completes, `GasUsed` is set and returned in the
  `abci.ExecTxResult`. If execution halts because `BlockGasMeter` or `GasMeter` has run out or something else goes
  wrong, a deferred function at the end appropriately errors or panics.

If there are any failed state changes resulting from a `Tx` being invalid or `GasMeter` running out,
the transaction processing terminates and any state changes are reverted. Invalid transactions in a
block proposal cause validator nodes to reject the block and vote for a `nil` block instead.

### Commit

The final step is for nodes to commit the block and state changes. Validator nodes
perform the previous step of executing state transitions in order to validate the transactions,
then sign the block to confirm it. Full nodes that are not validators do not
participate in consensus - i.e. they cannot vote - but listen for votes to understand whether or
not they should commit the state changes.

When they receive enough validator votes (2/3+ _precommits_ weighted by voting power), full nodes commit to a new block to be added to the blockchain and
finalize the state transitions in the application layer. A new state root is generated to serve as
a merkle proof for the state transitions. Applications use the [`Commit`](../advanced/00-baseapp.md#commit)
ABCI method inherited from [Baseapp](../advanced/00-baseapp.md); it syncs all the state transitions by
writing the `deliverState` into the application's internal state. As soon as the state changes are
committed, `checkState` starts afresh from the most recently committed state and `deliverState`
resets to `nil` in order to be consistent and reflect the changes.

Note that not all blocks have the same number of transactions and it is possible for consensus to
result in a `nil` block or one with none at all. In a public blockchain network, it is also possible
for validators to be **byzantine**, or malicious, which may prevent a `Tx` from being committed in
the blockchain. Possible malicious behaviors include the proposer deciding to censor a `Tx` by
excluding it from the block or a validator voting against the block.

At this point, the transaction lifecycle of a `Tx` is over: nodes have verified its validity,
delivered it by executing its state changes, and committed those changes. The `Tx` itself,
in `[]byte` form, is stored in a block and appended to the blockchain.
