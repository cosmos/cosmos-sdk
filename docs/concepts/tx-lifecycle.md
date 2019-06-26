# Transaction Lifecycle

## Prerequisite Reading

* [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This document describes the lifecycle of a transaction from creation to committed state changes. The transaction will be referred to as `Tx`.
1. [Creation](#creation)
2. [Addition to Mempool](#addition-to-mempool)
3. [Consensus](#consensus)
4. [State Changes](#state-changes)

## Creation

### Definition

Developers specify [**transactions**](./msg-tx.md#transactions), or actions that cause state changes in their applications by defining their corresponding [**messages**](./msg-tx.md#messages). Each transaction is comprised of metadata and one or multiple messages defined by developers by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/0c6d53dc077ee44ad72681b0bffafa1958f8c16d/types/tx_msg.go#L7-L31) interface.

### Transaction Creation

TODO: online vs offline

The transaction `tx` is created by the user inputting a command in the format `[modulename] tx [command] [args] [flags]` from the command-line, providing the type of transaction in `[command]`, arguments in `[args]`, and configurations such as gas prices in `[flags]`.

#### Gas and Fees

There are three flags the user can use to indicate how much he/she is willing to pay in fees:
* `--gas` specifies how much [gas](./fees-signature.md#gas), which represents computational resources, the transaction consumes. Gas is dependent on the transaction and is not precisely calculated until execution, but can be estimated by providing `auto` as the value for `--gas`, with an optional `--gas-adjustment` to scale up in order to avoid underestimating.  
* `--gas-prices` specifies how much the user is willing pay per unit of gas, which can be one or multiple denominations of tokens. For example, `--gas-prices=0.025uatom, 0.025upho` means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
* `-fees` specifies how much in fees the user is willing to pay in total.

The ultimate value of the fees paid is equal to the gas multiplied by the gas prices. In other words, `gas * gasPrices = fees`. Thus, since fees can be calculated using gas prices and vice versa, the user specifies only one of the two.

Later, validators may decide whether or not to include `tx` in their block depending on the fees or gas prices given. Generally, higher fees or gas prices earn higher priority, so users are incentivized to pay more.

#### Example
A user of the `gaia` application can input the following to generate a transaction to send 1000uatom from `sender-address` to `recipient-address`. This command will write the transaction, in JSON format, to the file `myUnsignedTx.json`.
```bash
gaiacli tx send <recipient-address> 1000uatom --from <sender-address> --gas auto -gas-prices 0.025uatom > myUnsignedTx.json
```

### Signature  
`Tx` must be signed by the relevant accounts owners in order to be valid, otherwise everyone could spend money from any account. To sign a transaction, the user enters a `sign` command.
Again, there are four values for flags that must be provided:

* `--from` specifies an address; the corresponding private key is used to sign the transaction.
* `--chain-id` specifies the unique identifier of the blockchain the transaction pertains to.
* `--sequence` is the value of a counter measuring how many transactions have been sent from the account. It is used to prevent replay attacks.
* `--account-number` is an identifier for the account.
* `--validate-signatures` (optional) instructs the process to sign and verify that all signatures have been provided.

Multisig transactions are supported: a multisig address needs to be created using the `keys add` command, and each signer signs the transaction serially.

#### Example
The following command signs the inputted transaction, `myUnsignedTx.json`, and writes the signed transaction to the file `mySignedTx.json`.
```bash
gaiacli tx sign myUnsignedTx.json --from <delegatorKeyName> --chain-id <chainId> --sequence <sequence> --account-number<accountNumber> > mySignedTx.json
```

### Broadcast

Up until broadcast, the transaction creation steps can be done offline (although may result in an invalid transaction if the user makes a mistake). The previously generated `Tx` is broadcasted to a peer node by the user entering a `broadcast` command. Only one flag is required here:
* `--node` specifies which node to broadcast to.
* `--trust-node` (optional) indicates whether or not the node and its response proofs can be trusted.
* `--broadcast-mode` (optional) specifies when the process should return. Options include asynchronous (return immediately), synchronous (return after `CheckTx` passes), or block (return after block commit).

#### Example
The following command broadcasts the signed transaction, `mySignedTx.json` to a particular node.
```bash
gaiacli tx broadcast mySignedTx.json --node <node>
```

## Addition to Mempool

<<<<<<< HEAD
Each full node that receives `tx` performs local checks to ensure it is not invalid. If approved, `tx` is held in the nodes' [**Mempool**](https://tendermint.com/docs/spec/reactors/mempool/functionality.html#external-functionality)s (memory pools unique to each node) pending approval from the rest of the network. Honest nodes will discard any transactions they find invalid. Prior to consensus, nodes continuously validate incoming transactions and gossip them to their peers.
=======
Each full node that receives `tx` performs local checks to ensure it is not invalid. If `tx` passes the checks, it is held in the nodes' [**Mempool**](https://github.com/tendermint/tendermint/blob/a0234affb6959a0aec285eebf3a3963251d2d186/state/services.go#L17-L34)s (memory pools unique to each node) pending inclusion in a block. Honest nodes will discard any transactions they find invalid. Prior to consensus, nodes continuously check incoming transactions and gossip them to their peers.
>>>>>>> links

### Types of Checks

**Stateless** checks do not require nodes to access state - light clients or offline nodes can do them. Stateless checks include making sure addresses are not empty, enforcing nonnegative numbers, and other logic specified in the `Msg` definition. Upon first receiving the transaction, `PreCheckFunc` can be called to reject transactions that are clearly invalid as early as possible, such as ones exceeding the block size. **Stateful** checks validate transactions and messages based on a committed state. After running stateless checks, nodes should also check that the relevant values exist and are able to be transacted with, the address has sufficient funds, and the sender is authorized or has the correct ownership to transact.

At a given moment, full nodes typically have multiple versions of the application's internal state for different purposes. For example, nodes will execute state changes while in the process of validating transactions, but still need a copy of the last committed state in order to answer queries - they should not respond using state that has changes not committed yet. The nodes' internal states start off at the most recent state agreed upon by the network, diverge as transactions are validated and executed, then re-sync after a new block's transactions are executed and committed.

The nodes validating `tx` call an ABCI validation function, `CheckTx` which includes both stateless and stateful checks:

### Decoding
`tx` is decoded. The [`runTx`](./baseapp.md#runtx) function is called to run in `runTxModeCheck` mode, meaning the function will not execute the messages.

<<<<<<< HEAD
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
### CheckTx
=======
### ValidateBasic
`Tx` is unpacked into its messages and [`validateBasic`](./msg-tx.md#validatebasic), a function required for every message to implement the `Msg` interface, is run for each one. It should include basic stateless sanity checks. For example, if the message is to send coins from one address to another, `validateBasic` likely checks for nonempty addresses and a nonnegative coin amount, but does not require knowledge of state such as account balance of an address.
>>>>>>> tx sign broadcast with cli

### AnteHandler
If an `AnteHandler` is defined, it is run. A deep copy of the internal state, `checkTxState`, is made and the `AnteHandler` performs the actions required for each message on it. Using a copy allows the handler to validate the transaction without modifying the last committed state, and revert back to the original if the execution fails.

### Gas
The `Context` used to keep track of important data while `AnteHandler` is executing `tx` keeps a `GasMeter` which tracks how much gas has been used.
Gas is known as the value `GasWanted`. `CheckTx` returns `GasUsed` which may or may not be less than the user's provided `GasWanted`. A `PostCheckFunc` is an optional filter run after `CheckTx` that can be used to enforce that users provide enough gas.

If at any point during the checking stage `tx` is found to be invalid, it is discarded and the transaction lifecycle ends here. Otherwise, the default protocol is to add it to the Mempool, and `tx` becomes a candidate to be included in the next block. Note that even if `tx` passes all checks at this stage, it is still possible to be found invalid later on as these checks are limited.


## Consensus

Consensus happens in **rounds**: each round begins with a proposer creating a block of the most recent transactions and ends with validators agreeing to accept the block or go with a nil block instead. One proposer is chosen amongst the validators to create and propose their chosen block. In the previous step, this validator generated a Mempool of validated transactions, including `tx`, and now includes them in a block.

The validators execute [Tendermint BFT Consensus](https://tendermint.com/docs/spec/consensus/consensus.html#terms); with +2/3 commits from the validators, the block is committed. Note that not all blocks have the same number of transactions and it is possible for consensus to result in a nil block or one with no transactions.

In a public blockchain network, it is possible for validators to be **byzantine**, or malicious, which may prevent `tx` from being committed in the blockchain. Possible malicious behaviors include the proposer deciding to censor the transaction by excluding it from the block or a validator voting against the block.

If `tx` is included in a block that the network commits to, it officially becomes part of the blockchain, which logs the application's transaction history.

## State Changes

During consensus, the validators came to agreement on not only which transactions but also the precise order of the transactions. However, apart from committing to this block in consensus, the ultimate goal is actually for nodes to commit to a new state generated by the transaction state changes.
In order to execute `tx`, the following ABCI function calls are made in order to communicate to the application what state changes need to be made. While nodes each run everything individually, since the messages' state transitions are deterministic and the order was finalized during consensus, this process yields a single, unambiguous result.
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
### Block Functions

There are a few other functions that the full-nodes run to commit the block.
[`BeginBlock`](./app-anatomy.md#beginblocker-and-endblocker) is run before `DeliverTx`, and
[`EndBlock`](./app-anatomy.md#beginblocker-and-endblocker) is always run afterward.

### DeliverTx

The `DeliverTx` ABCI function defined in [`baseapp`](./baseapp.md) does the bulk of the state change work: it is run for each transaction in the block in sequential order as committed to during consensus. Under the hood, `DeliverTx` is almost identical to `CheckTx` but calls the [`runTx`](./baseapp.md#runtx) function in deliver mode instead of check mode. Instead of using their `checkTxState` or `queryState`, nodes select a new copy, `deliverTxState`, to deliver transactions:

* **Decoding:** `Tx` is decoded.
* **Initializations:** The `Context`, `Multistore`, and `GasMeter` are initialized.
* **Checks:** The nodes call `validateBasicMsgs` and the `AnteHandler` again. This second check happens because nodes may not have seen the same transactions during the Addition to Mempool stage and a malicious proposer may have included invalid transactions.
* **Route and Handler:** The extra step is to run `runMsgs` to fully execute each `Msg` within the transaction. Since `tx` may have messages from different modules, `baseapp` needs to know which module to find the appropriate Handler. Thus, the [`Route`](./msg-tx.md#route) function is called to retrieve the route name and find the `MsgHandler`.
* **MsgHandler:** The `MsgHandler`, a step up from `AnteHandler`, is responsible for executing each message's actions and causes state changes to persist in `deliverTxState`. It is defined within a `Msg`'s module and writes to the appropriate stores within the module.
* **Gas:** While the transaction is being delivered, a `GasMeter` is used to keep track of how much gas is left for each transaction; GasUsed is deducted from it and returned in the `Response`. Deferred functions run at the end of `runTx` for the purpose of checking how much gas was used and outputting `GasUsed` and `GasWanted` in the `Response`. If `GasMeter` has run out or something else goes wrong, it appropriately errors or panics.

If there are any failed state changes resulting from `tx` being invalid or `GasMeter` running out, `tx` processing terminates and any state changes are reverted. Any leftover gas is returned to the user.

### EndBlock



#### Commit

The application's `Commit` method is run in order to finalize the state changes made by executing this block's transactions. A new state root should be sent back to serve as a merkle proof for the state change. Any application can inherit Baseapp's [`Commit`](./baseapp.md#commit) method; it syncs all the states by writing the `deliverTxState` into the application's internal state. Moving forward, `checkTxState`, `deliverTxState`, and `queryState` will all start afresh from the most recently committed state in order to be consistent and reflect the changes.

## Next

Learn about [Baseapp](./baseapp.md).
