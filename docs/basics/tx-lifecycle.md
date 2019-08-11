# Transaction Lifecycle

## Prerequisite Reading

* [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This document describes the lifecycle of a transaction from creation to committed state changes. Transaction definition is described in a [different doc](../core/tx-msgs.md#transactions). The transaction will be referred to as `Tx`.

1. [Creation](#creation)
2. [Addition to Mempool](#addition-to-mempool)
3. [Inclusion in a Block](#inclusion-in-a-block)
4. [State Changes](#state-changes)
5. [Consensus and Commit](#consensus-and-commit)

## Creation

### Transaction Creation

One of the main application interfaces is the command-line interface. The transaction `Tx` can be created by the user inputting a command in the following format from the [command-line](../interfaces/cli.md), providing the type of transaction in `[command]`, arguments in `[args]`, and configurations such as gas prices in `[flags]`:

```bash
[appname] tx [command] [args] [flags]
```

This command will automatically **create** the transaction, **sign** it using the account's private key, and **broadcast** it to the specified peer node.

There are several required and optional flags for transaction creation. The `--from` flag specifies which [account](./accounts-fees-gas.md#accounts) the transaction is orginating from. For example, if the transaction is sending coins, the funds will be drawn from the specified `from` address.

#### Gas and Fees

Additionally, there are several [flags](../interfaces/cli.md) users can use to indicate how much they are willing to pay in [fees](./accounts-fees-gas.md#fees):

* `--gas` refers to how much [gas](./accounts-fees-gas.md#gas), which represents computational resources, `Tx` consumes. Gas is dependent on the transaction and is not precisely calculated until execution, but can be estimated by providing `auto` as the value for `--gas`.
* `--gas-adjustment` (optional) can be used to scale `gas` up in order to avoid underestimating. For example, users can specify their gas adjustment as 1.5 to use 1.5 times the estimated gas.
* `--gas-prices` specifies how much the user is willing pay per unit of gas, which can be one or multiple denominations of tokens. For example, `--gas-prices=0.025uatom, 0.025upho` means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
* `--fees` specifies how much in fees the user is willing to pay in total.

The ultimate value of the fees paid is equal to the gas multiplied by the gas prices. In other words, `fees = ceil(gas * gasPrices)`. Thus, since fees can be calculated using gas prices and vice versa, the users specify only one of the two.

Later, validators decide whether or not to include the transaction in their block by comparing the given or calculated `gas-prices` to their local `min-gas-prices`. `Tx` will be rejected if its `gas-prices` is not high enough, so users are incentivized to pay more.

#### CLI Example

Users of application `app` can enter the following command into their CLI to generate a transaction to send 1000uatom from a `senderAddress` to a `recipientAddress`. It specifies how much gas they are willing to pay: an automatic estimate scaled up by 1.5 times, with a gas price of 0.025uatom per unit gas.

```bash
appcli tx send <recipientAddress> 1000uatom --from <senderAddress> --gas auto --gas-adjustment 1.5 --gas-prices 0.025uatom
```

#### Other Transaction Creation Methods

The command-line is an easy way to interact with an application, but `Tx` can also be created using a [REST interface](../interfaces/rest.md) or some other entrypoint defined by the application developer. From the user's perspective, the interaction depends on the web interface or wallet they are using (e.g. creating `Tx` using [Lunie.io](https://lunie.io/#/) and signing it with a Ledger Nano S).

## Addition to Mempool

Each full-node (running Tendermint) that receives `Tx` sends an [ABCI message](https://tendermint.com/docs/spec/abci/abci.html#messages), `CheckTx`, to the application layer to check for invalidity, and receives a Response. If `Tx` passes the checks, it is held in the nodes' [**Mempool**](https://tendermint.com/docs/tendermint-core/mempool.html#mempool)s (memory pools of transactions unique to each node) pending inclusion in a block - honest nodes will discard `Tx` if it is found to be invalid. Prior to consensus, nodes continuously check incoming transactions and gossip them to their peers.


### Types of Checks

The full-nodes perform stateless, then stateful checks on `Tx` during `CheckTx`, with the goal to identify and reject an invalid transaction as early on as possible to avoid wasted computation. 

***Stateless*** checks do not require nodes to access state - light clients or offline nodes can do them - and are thus less computationally expensive. Stateless checks include making sure addresses are not empty, enforcing nonnegative numbers, and other logic specified in the definitions.

***Stateful*** checks validate transactions and messages based on a committed state. Examples include checking that the relevant values exist and are able to be transacted with, the address has sufficient funds, and the sender is authorized or has the correct ownership to transact. At any given moment, full-nodes typically have [multiple versions](../core/baseapp.md#volatile-states) of the application's internal state for different purposes. For example, nodes will execute state changes while in the process of verifying transactions, but still need a copy of the last committed state in order to answer queries - they should not respond using state with uncommitted changes.

In order to verify `Tx`, full-nodes call `CheckTx`, which includes both _stateless_ and _stateful_ checks. Further validation happens later in the [`DeliverTx`](#delivertx) stage. `CheckTx` goes through several steps, beginning with decoding `Tx`.

### Decoding

When `Tx` is received by the application from the underlying consensus engine (e.g. Tendermint), it is still in its encoded (i.e. using [Amino](https://tendermint.com/docs/spec/blockchain/encoding.html#amino)) `[]byte` form and needs to be unmarshaled in order to be processed. Then, the [`runTx`](../core/baseapp.md#runtx-and-runmsgs) function is called to run in `runTxModeCheck` mode, meaning the function will run all checks but exit before executing messages and writing state changes.

### ValidateBasic

[Messages](../core/tx-msgs.md#messages) are extracted from `Tx` and `ValidateBasic`, a function defined by the module developer for every message, is run for each one. It should include basic stateless sanity checks. For example, if the message is to send coins from one address to another, `ValidateBasic` likely checks for nonempty addresses and a nonnegative coin amount, but does not require knowledge of state such as account balance of an address.

### AnteHandler

The [`AnteHandler`](../core/baseapp.md#antehandler), which is technically optional but should be defined for each application, is run. A deep copy of the internal state, `checkState`, is made and the defined `AnteHandler` performs limited checks specified for the transaction type. Using a copy allows the handler to do stateful checks for `Tx` without modifying the last committed state, and revert back to the original if the execution fails.

For example, the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/master/docs/spec/auth) module `AnteHandler` checks and increments sequence numbers, checks signatures and account numbers, and deducts fees from the first signer of the transaction - all state changes are made using the `checkState`.

### Gas

The [`Context`](../core/context.md) used to keep track of important data while `AnteHandler` is executing `Tx` keeps a `GasMeter` which tracks how much gas has been used. The user-provided amount for gas is known as the value `GasWanted`. If `GasConsumed`, the amount of gas consumed so far, ever exceeds `GasWanted`, execution stops. Otherwise, `CheckTx` sets `GasUsed` equal to `GasConsumed` and returns it in the result. After calculating the gas and fee values, validator-nodes check that the user-specified `gas-prices` is less than their locally defined `min-gas-prices`.

### Discard or Addition to Mempool

If at any point during the checking stage `Tx` fails a check, it is discarded and the transaction lifecycle ends here. Otherwise, if it passes `CheckTx` successfully, the default protocol is to relay it to peer nodes and add it to the Mempool so that `Tx` becomes a candidate to be included in the next block.

The **mempool** serves the purpose of keeping track of transactions seen by all full-nodes. Full-nodes that are not validators (and thus do not participate in creating blocks) keep a **mempool cache** of the last `[mempool] cache_size` transactions they have seen, as a first line of defense to prevent replay attacks. Ideally, `[mempool] cache_size` is large enough to encompass all of the transactions in the full mempool. If the the mempool cache is too small to keep track of all the transactions, `CheckTx` is responsible for identifying and rejecting replayed transactions. Currently existing preventative measures include fees and a `sequence` counter to distinguish replayed transactions from identical but valid ones. If an attacker tries to spam nodes with many copies of `Tx`, full-nodes keeping a mempool cache will reject identical copies instead of running `CheckTx` on all of them. Even if the copies have incremented `sequence` numbers, attackers are disincentivized by the need to pay fees.

Validator-nodes keep a mempool to prevent replay attacks, just as full-nodes do, but also use it as a pool of unconfirmed transactions in preparation to include them in a block. Note that even if `Tx` passes all checks at this stage, it is still possible to be found invalid later on, because `CheckTx` does not fully validate the transaction.

## Inclusion in a Block

Consensus, the process through which validator-nodes come to agreement on which transactions to accept, happens in **rounds**. Each round begins with a proposer creating a block of the most recent transactions and ends with **validators**, special full-nodes with voting power responsible for consensus, agreeing to accept the block or go with a `nil` block instead. Validator-nodes execute the consensus algorithm, such as [Tendermint BFT](https://tendermint.com/docs/spec/consensus/consensus.html#terms), confirming the transactions using ABCI calls to the application, in order to come to this agreement.

The first step of consensus is the **block proposal**. One proposer amongst the validators is chosen by the consensus algorithm to create and propose a block - in order for `Tx` to be included, it must be in this proposer's mempool.

## State Changes

The next step of consensus is to execute the transactions to fully validate them. All full-nodes that receive a block proposal execute the transactions by calling the ABCI functions [`BeginBlock`](./app-anatomy.md#beginblocker-and-endblocker), `DeliverTx` for each transaction, and [`EndBlock`](./app-anatomy.md#beginblocker-and-endblocker). While full-nodes each run everything individually, since the messages' state transitions are deterministic and transactions are explicitly ordered in the block proposal, this process yields a single, unambiguous result.

```
		-----------------------		
		|Receive Block Proposal|        
		-----------------------		
		          |		
			  v
		-----------------------		
		| BeginBlock	      |        
		-----------------------		
		          |		
			  v			
		-----------------------		    
		| DeliverTx(tx0)      |  
		| DeliverTx(tx1)      |   	  
		| DeliverTx(tx2)      |  
		| DeliverTx(tx3)      |  
		|	.	      |  
		|	.	      |
		|	.	      |
		-----------------------		
		          |			
			  v			
		-----------------------
		| EndBlock	      |         
		-----------------------
		          |			
			  v			
		-----------------------
		| Consensus	      |         
		-----------------------
		          |			
			  v			
		-----------------------
		| Commit	      |         
		-----------------------
```

### DeliverTx

The `DeliverTx` ABCI function defined in [`baseapp`](../core/baseapp.md) does the bulk of the state change work: it is run for each transaction in the block in sequential order as committed to during consensus. Under the hood, `DeliverTx` is almost identical to `CheckTx` but calls the [`runTx`](../core/baseapp.md#runtx-and-runmsgs) function in deliver mode instead of check mode. Instead of using their `checkState` or `queryState`, full-nodes select a new copy, `deliverState`, to deliver `Tx`:

* **Decoding:** Since `DeliverTx` is an ABCI call, `Tx` is received in the encoded `[]byte` form. Nodes first unmarshal the transaction, then call `runTx` in `runTxModeDeliver`, which is very similar to `CheckTx` but also executes and writes state changes.
* **Checks:** Full-nodes call `validateBasicMsgs` and the `AnteHandler` again. This second check happens because they may not have seen the same transactions during the Addition to Mempool stage and a malicious proposer may have included invalid ones. One difference here is that the `AnteHandler` will not compare `gas-prices` to the node's `min-gas-prices`since that value is local to each node - differing values across nodes would yield nondeterministic results.
* **Route and Handler:** While `CheckTx` would have exited, `DeliverTx` continues to run [`runMsgs`](../core/baseapp.md#runtx-and-runmsgs) to fully execute each `Msg` within the transaction. Since the transaction may have messages from different modules, `baseapp` needs to know which module to find the appropriate Handler. Thus, the [`Route`](../core/tx-msgs.md#route)
 function is called to retrieve the route name and find the `Handler` within the module.
* **Handler:** The `Handler`, a step up from `AnteHandler`, is responsible for executing each message's actions and causes state changes to persist in `deliverTxState`. It is defined within a `Msg`'s module and writes to the appropriate stores within the module.
* **Gas:** While `Tx` is being delivered, a `GasMeter` is used to keep track of how much gas is left for each transaction; if execution completes, `GasUsed` is set and returned in the `Response`. If execution halts because `GasMeter` has run out or something else goes wrong, a deferred function at the end appropriately errors or panics.

If there are any failed state changes resulting from `Tx` being invalid or `GasMeter` running out, the transaction processing terminates and any state changes are reverted. Invalid transactions in a block proposal cause validator-nodes to reject the block and vote for a `nil` block instead. If `Tx` is delivered successfully, any leftover gas is returned to the user and the transaction is validated.

### Consensus and Commit

The final step is for validator-nodes participating in consensus to commit the block and state changes. Validator-nodes perform the previous step of executing state changes in order to validate the transactions, then sign the block to confirm and vote for them. Full-nodes that are not validators do not participate in consensus - i.e. they cannot vote - but listen for votes to understand whether or not they should commit the state changes.

#### Consensus

The consensus layer may technically run any consensus algorithm to come to agreement on which block to accept. In [Tendermint BFT](https://tendermint.com/docs/spec/consensus/consensus.html), validator-nodes go through a Prevote stage and a Precommit stage, and the block requires +2/3 Precommits from the validator-nodes to move into the commit stage. Note that this phase is purely in the consensus layer; `Tx` is represented as a `[]byte`.

#### Commit

In the **Commit** stage, full-nodes commit to a new block to be added to the blockchain and finalize the state changes on the application layer. A new state root is generated to serve as a merkle proof for the state change. Applications use the [`Commit`](../core/baseapp.md#commit) ABCI method inherited from [Baseapp](../core/baseapp.md); it syncs all the states by writing the `deliverState` into the application's internal state. As soon as the state changes are committed, `checkState` and `queryState` start afresh from the most recently committed state and `deliverState` resets to `nil` in order to be consistent and reflect the changes.

Note that not all blocks have the same number of transactions and it is possible for consensus to result in a `nil` block or one with none at all. In a public blockchain network, it is also possible for validators to be **byzantine**, or malicious, which may prevent `Tx` from being committed in the blockchain. Possible malicious behaviors include the proposer deciding to censor `Tx` by excluding it from the block or a validator voting against the block.

At this point, the transaction lifecycle of `Tx` is over: nodes have verified its validity, delivered it by executing its state changes, and committed those changes. The `Tx` itself, in `[]byte` form, is stored in a block and appended to the blockchain.

## Next

Learn about [Baseapp](../core/baseapp.md).
