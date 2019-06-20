# Transaction Lifecycle

## Prerequisite Reading

* [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This document describes the lifecycle of a transaction from creation to committed state changes. The transaction will be referred to as `tx`.
1. [Creation](#creation)
2. [Addition to Mempool](#addition-to-mempool)
3. [Consensus](#consensus)
4. [State Changes](#state-changes)

## Creation

### Definition

Developers specify [**transactions**](./msg-tx.md#transactions), or actions that cause state changes in their applications by defining their corresponding [**messages**](./msg-tx.md#messages). Each transaction is comprised of metadata and one or multiple messages defined by developers by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/0c6d53dc077ee44ad72681b0bffafa1958f8c16d/types/tx_msg.go#L7-L31) interface.

Each application may include one or more **modules** that compartmentalize the application's capabilities. Some are provided with the SDK (such as `auth` for signatures and `bank` for transferring value); developers will also define their own. Message are each housed in a specific module and thus include a `Route()` function to route to the correct module.

### User Creation

The transaction `tx` is created by running `appcli tx [tx]` from the command-line, providing transaction data in `[tx]` and, optionally, configurations such as fees or gas, broadcast mode, and whether to only generate offline by appending flags at the end.

Transaction senders may supply **fees** (similar to transaction fees in Bitcoin) or **gas prices** (similar to gas in Ethereum) using the `--fees` or `--gas-prices` flags, respectively. Note that only one of the two can be used. Later, validators may decide whether or not to include `tx` in their block depending on the fees or gas prices given. Generally, higher fees or gas prices generally earn higher priority, but the senders are also able to indicate the maximum fees or gas they are willing to pay.

For example, the following command creates a `send` transaction where the user is willing to pay 0.025uatom per unit gas.
```bash
gaiacli tx send [from_key_or_address] [to_address] [amount] --gas-prices=0.025uatom
```

### Subcommands

The SDK uses [Cobra Commands](https://godoc.org/github.com/spf13/cobra#Command) to instruct nodes what to do. The command called for `tx` is `txCmd`, which includes one or more subcommands that depend on the application's functionalities and what `tx` is defined to do. The following are common subcommands, some already defined:
* `SendTxCmd` generates a command to send value from one address to another. Sending value is very common but not all transactions are strictly financial (state changes of many types are possible).
* `GetSignCommand` generates a command to sign the transaction and prints the JSON encoding of the transaction. It may also output the JSON encoding of the generated signature. If the --validate-signatures flag is toggled on, it also verifies that the signers required for the transaction have provided valid signatures in the correct order.
* `GetEncodeCommand` creates a command that uses the [Amino]() format to encode the transaction.
* `GetBroadcastCommand` creates the Tx Broadcast command used to share the transaction, inputted as a JSON file, with a full node.
* Other commands defined by the application developer.

### Broadcast

Up until broadcast, the transaction creation steps can be done offline (although may result in an invalid transaction). It is possible for the transaction to be created, then signed, then broadcasted using separate commands. The previously generated Tx Broadcast command is generated within the SDK, then executed by the originating node running Tendermint Core. `Tx` is broadcasted at this layer in the generic `[]byte` form.

## Addition to Mempool

Each full node that receives `tx` performs local checks to ensure it is not invalid. If `tx` passes the checks, it is held in the nodes' [**Mempool**](https://tendermint.com/docs/tendermint-core/mempool.html#mempool)s (memory pools unique to each node) pending inclusion in a block. Honest nodes will discard any transactions they find invalid. Prior to consensus, nodes continuously check incoming transactions and gossip them to their peers.

### Internal State

**State** encompasses all relevant data in an application and is what the state machine stores. It includes data such as account balances, ownership of assets, and other relevant information that may change over time with executed transactions. State is stored in [Stores](./store/README.md).

**Stateless** checks do not require nodes to access state - light clients or offline nodes can do them. Stateless checks include making sure addresses are not empty, enforcing nonnegative numbers, and other logic specified in the `Msg` definition. Upon first receiving the transaction, `PreCheckFunc` can be called to reject transactions that are clearly invalid as early as possible, such as ones exceeding the block size. **Stateful** checks validate transactions and messages based on a committed state. After running stateless checks, nodes should also check that the relevant values exist and are able to be transacted with, the address has sufficient funds, and the sender is authorized or has the correct ownership to transact.

At a given moment, full nodes typically have multiple versions of the application's internal state for different purposes. For example, nodes will execute state changes while in the process of validating transactions, but still need a copy of the last committed state in order to answer queries - they should not respond using state that has changes not committed yet. The nodes' internal states start off at the most recent state agreed upon by the network, diverge as transactions are validated and executed, then re-sync after a new block's transactions are executed and committed.


### CheckTx

The nodes validating `tx` call an ABCI validation function, `CheckTx` which includes both stateless and stateful checks:
* **Decoding:** Interfacing with the application, `tx` is first decoded.
* **RunTx:** The [`runTx`](./baseapp.md#runtx) function is called to run in `runTxModeCheck` mode, meaning the function will not execute the messages.
* **ValidateBasic:** `Tx` is unpacked into its messages and [`validateBasic`](./msg-tx.md#validatebasic), a function required for every message to implement the `Msg` interface, is run for each one. It should include basic stateless sanity checks. For example, if the message is to send coins from one address to another, `validateBasic` likely checks for nonempty addresses and a nonnegative coin amount, but does not require knowledge of state such as account balance of an address.
* **AnteHandler:** If an `AnteHandler` is defined, it is run. A deep copy of the internal state, `checkTxState`, is made and the `AnteHandler` performs the actions required for each message on it. Using a copy allows the handler to validate the transaction without modifying the last committed state, and revert back to the original if the execution fails.
* **Gas:** The `Context` used to keep track of important data while `AnteHandler` is executing `tx` keeps a `GasMeter` which tracks how much gas has been used.
* **Response:** `RunTx` returns a result, which `CheckTx` formats into an ABCI `Response` which includes a log, data pertaining to the messages involved, and information about the amount of gas used.

### Gas Checking

If the transaction sender inputted maximum gas previously, it is known as the value `GasWanted`. `CheckTx` returns `GasUsed` which may or may not be less than the user's provided `GasWanted`. A `PostCheckFunc` is an optional filter run after `CheckTx` that can be used to enforce that users provide enough gas.

If at any point during the checking stage `tx` is found to be invalid, the default protocol is to discard it and the transaction's lifecycle ends here. Otherwise, the default protocol is to add it to the Mempool, and `tx` becomes a candidate to be included in the next block. Note that even if `tx` passes all checks at this stage, it is still possible to be found invalid later on as these checks are limited.

## Consensus

Consensus happens in **rounds**: each round begins with a proposer creating a block of the most recent transactions and ends with **validators**, special full-nodes with voting power responsible for consensus, agreeing to accept the block or go with a nil block instead. One proposer is chosen amongst the validators to create and propose a block. In the previous step, this validator generated a Mempool of validated transactions, including `tx`, and now includes them in a block.

The validators execute [Tendermint BFT Consensus](https://tendermint.com/docs/spec/consensus/consensus.html#terms); with +2/3 precommit votes from the validators, the block is committed. Note that not all blocks have the same number of transactions and it is possible for consensus to result in a nil block or one with no transactions.

### Byzantine Validators

In a public blockchain network, it is possible for validators to be **byzantine**, or malicious, which may prevent `tx` from being committed in the blockchain. Some malicious behaviors are detectable, such as going offline (unavailable to vote) or voting on multiple blocks in one round - these behaviors are logged and can be penalized accordingly. Others are less obvious, such as the proposer deciding to censor the transaction by excluding it from the block.

If `tx` is included in a block that the network commits to, it officially becomes part of the blockchain, which logs the application's transaction history.

## State Changes

During consensus, all full-nodes come to agreement on not only which transactions but also the precise order of the transactions in their block. However, apart from committing to this block in consensus, the ultimate goal is actually for full-nodes to commit to a new state generated by the transaction state changes.
In order to execute `tx`, the following ABCI function calls are made in order to communicate to the application what state changes need to be made. While full-nodes each run everything individually, since the messages' state transitions are deterministic and the order was finalized during consensus, this process yields a single, unambiguous result.
```
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
		| Commit	      |         
		-----------------------
```
#### BeginBlock

[`BeginBlock`](./app-anatomy.md#beginblocker-and-endblocker) is run first, and mainly transmits important data such as block header and Byzantine Validators from the last round of consensus to be used during the next few steps. No transactions are handled here.

#### DeliverTx

The `DeliverTx` ABCI function defined in [`baseapp`](./baseapp.md) does the bulk of the state change work: it is run for each transaction in the block in sequential order as committed to during consensus. Under the hood, `DeliverTx` is almost identical to `CheckTx` but calls the [`runTx`](./baseapp.md#runtx) function in deliver mode instead of check mode. Instead of using their `checkTxState` or `queryState`, full-nodes select a new copy, `deliverTxState`, to deliver transactions:

* * **Decoding:** `Tx` is decoded.
* **Initializations:** The `Context`, `Multistore`, and `GasMeter` are initialized.
* **Checks:** The full-nodes call `validateBasicMsgs` and the `AnteHandler` again. This second check happens because they may not have seen the same transactions during the Addition to Mempool stage and a malicious proposer may have included invalid transactions.
* **Route and Handler:** The extra step is to run `runMsgs` to fully execute each `Msg` within the transaction. Since `tx` may have messages from different modules, `baseapp` needs to know which module to find the appropriate Handler. Thus, the [`Route`](./msg-tx.md#route) function is called to retrieve the route name and find the `MsgHandler`.
* **MsgHandler:** The `MsgHandler`, a step up from `AnteHandler`, is responsible for executing each message's actions and causes state changes to persist in `deliverTxState`. It is defined within a `Msg`'s module and writes to the appropriate stores within the module.
* **Gas:** While the transaction is being delivered, a `GasMeter` is used to keep track of how much gas is left for each transaction; GasUsed is deducted from it and returned in the `Response`. Deferred functions run at the end of `runTx` for the purpose of checking how much gas was used and outputting `GasUsed` and `GasWanted` in the `Response`. If `GasMeter` has run out or something else goes wrong, it appropriately errors or panics.

If there are any failed state changes resulting from `tx` being invalid or `GasMeter` running out, `tx` processing terminates and any state changes are reverted. Any leftover gas is returned to the user.

#### EndBlock

[`EndBlock`](./app-anatomy.md#beginblocker-and-endblocker) is always run at the end and is useful for automatic function calls or changing governance/validator parameters. No transactions are handled here.

#### Commit

The application's `Commit` method is run in order to finalize the state changes made by executing this block's transactions. A new state root should be sent back to serve as a merkle proof for the state change. Any application can inherit Baseapp's [`Commit`](./baseapp.md#commit) method; it syncs all the states by writing the `deliverTxState` into the application's internal state. Moving forward, `checkTxState`, `deliverTxState`, and `queryState` will all start afresh from the most recently committed state in order to be consistent and reflect the changes.

Note that `CheckTx` and `DeliverTx` are called for every transaction, while `Commit` is called once for the whole block:

```
		To perform state checks	  	   To execute state 		   To answer queries
		on received transactions       transitions during Commit	about last-committed state

		-----------------------		-----------------------		-----------------------
		| CheckTxState(t)(0)  |         | DeliverTxState(t)(0)|		|    QueryState(t)    |
		-----------------------		|                     |		|                     |
CheckTx()	          |			|                     |		|                     |
			  v			|                     |		|                     |
		-----------------------		|                     |		|                     |
		| CheckTxState(t)(1)  |         |                     |		|                     |
		-----------------------		|                     |		|                     |
CheckTx()	          |			|                     |		|                     |
			  v			|                     |		|                     |
		-----------------------		|                     |		|                     |
		| CheckTxState(t)(2)  |         |   		      |		|                     |
		-----------------------		|                     |		|                     |
CheckTx()	          |			|                     |		|                     |
			  v			|                     |		|                     |
		-----------------------		|                     |		|                     |
		| CheckTxState(t)(3)  |         -----------------------		|                     |
DeliverTx()	|   		      |		           |          		|                     |
		|   		      |		           v          		|                     |
		|   		      |		-----------------------		|                     |
		|   		      |		| DeliverTxState(t)(1)|		|                     |
		|   		      |		-----------------------		|                     |
DeliverTx()	|   		      |	                   |			|                     |
		|   		      |			   v			|                     |
		|   		      |		-----------------------		|                     |
		|   		      |	      	| DeliverTxState(t)(2)|		|                     |
		|   		      |		-----------------------		|                     |
DeliverTx()	|   		      |	                   |			|                     |
		|   		      |			   v			|                     |
		|   		      |		-----------------------		|                     |
		|   		      |	      	| DeliverTxState(t)(3)|		|                     |
		-----------------------		-----------------------		-----------------------
Commit()		  |				   |				   |
			  v				   v				   v
		-----------------------		-----------------------		-----------------------
States synced:  | CheckTxState(t+1)   |         | DeliverTxState(t+1) |		|   QueryState(t+1)   |
		-----------------------		|                     |		|		      |
		          |			|                     |		|		      |
			  v			|                     |		|	  	      |
			  .				   .				   .
			  .				   . 				   .
			  .				   .  				   .

```

## Next

Learn about [Baseapp](./baseapp.md).
