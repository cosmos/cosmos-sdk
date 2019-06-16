# Transaction Lifecycle

## Prerequisite Reading
* [Anatomy of an SDK Application](https://github.com/cosmos/cosmos-sdk/blob/master/docs/concepts/app-anatomy.md)

## Synopsis
This document describes the lifecycle of a transaction from creation to committed state changes. The transaction will be referred to as `tx`.
1. [Creation](https://github.com/cosmos/cosmos-sdk/blob/master/docs/concepts/tx-lifecycle.md#creation)
2. [Addition to Mempool](https://github.com/cosmos/cosmos-sdk/blob/master/docs/concepts/tx-lifecycle.md#addition-to-mempool)
3. [Consensus](https://github.com/cosmos/cosmos-sdk/blob/master/docs/concepts/tx-lifecycle.md#consensus)
4. [State Changes](https://github.com/cosmos/cosmos-sdk/blob/master/docs/concepts/tx-lifecycle.md#state-changes)

## Creation
### Definition
Developers will specify [transactions](https://github.com/cosmos/cosmos-sdk/blob/master/docs/concepts/msg-tx.md#transactions), or actions that cause state changes in their applications. Each transaction is comprised of metadata and one or multiple [messages](https://github.com/cosmos/cosmos-sdk/blob/master/docs/concepts/msg-tx.md#messages) which are also defined by developers by implementing the `[Msg](https://github.com/cosmos/cosmos-sdk/blob/0c6d53dc077ee44ad72681b0bffafa1958f8c16d/types/tx_msg.go#L7-L31)` interface.

### User Creation
Transaction senders can create a transaction by running `appcli tx [tx]` from the command-line, providing transaction data and, optionally, configurations such as fees or gas, broadcast mode, and whether to only generate offline.

Transaction senders may supply **fees** (similar to transaction fees in Bitcoin) or **gas prices** (similar to gas in Ethereum) using the `--fees` or `--gas-prices` flags, respectively. Note that only one of the two can be used. Later, validators may decide which transactions to include in their block depending on the fees or gas prices given. Generally, higher fees or gas prices generally earn higher priority, but the senders are also able to indicate the maximum fees or gas they are willing to pay.

For example, the following command creates a `send` transaction where the user is willing to pay 0.025uatom per unit gas.
```bash
gaiacli tx send [from_key_or_address] [to_address] [amount] --gas-prices=0.025uatom
````
More [flags](https://github.com/cosmos/cosmos-sdk/blob/8c89023e9f7ce67492142c92acc9ba0d9f876c0e/client/flags/flags.go#L15-L58).
### Subcommands
The command called is `txCmd`, which includes several subcommands that depend on the application's functionalities:
* `SendTxCmd` is a command created from the `bank` module, used to send value from one address to another. Sending value is very common but not all transactions are strictly financial (state changes of many types are possible).
* `GetSignCommand` is created from the `auth` module, generating a command to sign the transaction and prints the JSON encoding of the transaction. It may also output the JSON encoding of the generated signature. If the --validate-signatures flag is toggled on, it also verifies that the signers required for the transaction have provided valid signatures in the correct order.
* `GetEncodeCommand` uses the Amino format to encode the transaction.
* `GetBroadcastCommand` is used to share the transaction, inputted as a JSON file, with a full node.
* Other commands defined by the application developer.

### Broadcast
Up until broadcast, the transaction creation steps can be done offline (although may result in an invalid transaction).

## Addition to Mempool
Each full node that receives a transaction performs local checks to filter out invalid transactions before they get included in a block. The transactions approved by a node are held in its [**Mempool**](https://github.com/tendermint/tendermint/blob/75ffa2bf1c7d5805460d941a75112e6a0a38c039/mempool/mempool.go) (memory pool unique to each node) pending approval from the rest of the network. Honest nodes will discard any transactions they find invalid. Prior to consensus, nodes continuously validate incoming transactions and gossip them to their peers.

### Internal State
**Stateless** checks do not require the node to access the previous state - a light client or offline node can do them. Stateless checks include making sure addresses are not empty, enforcing nonnegative numbers, and other logic specified in the `Msg` definition. Upon first receiving the transaction, `PreCheckFunc` can be called to reject transactions that are clearly invalid as early as possible, such as ones exceeding the block size. **Stateful** checks validate transactions and messages based on a committed state. After running stateless checks, nodes should also check that the relevant values exist and are able to be transacted with, the address has sufficient funds, and the sender is authorized or has the correct ownership to transact.

At a given moment, full nodes typically have multiple versions of the application's internal state for different purposes. For example, nodes will execute state changes while in the process of validating transactions, but still need a copy of the last committed state in order to answer queries - they should not respond using state that has not been committed yet. The nodes' internal states start off at the most recent state agreed upon by the network, diverge as transactions are validated and executed, then re-sync after a new block's transactions are executed and committed.


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
The nodes validating transactions call an ABCI validation function, `CheckTx` to run stateful checks:
* Nodes running Tendermint broadcast transactions in the generic `[]byte` form. Interfacing with the application, the transaction is first decoded.
* The `runTx` function is called to run in `runTxModeCheck` mode, meaning the function will not execute the messages.
* The transaction is unpacked into its messages and `validateBasic` is run for each one.
* If an `AnteHandler` is defined, it is run. A deep copy of the internal state, `checkTxState`, is made and the `AnteHandler` performs the actions required for each message on it. Using a copy allows the handler to validate the transaction without modifying the last committed state, and revert back to the original if the execution fails.  The stateful check is able to detect errors such as insufficient funds held by an address.
* The `[Context]`(https://github.com/cosmos/cosmos-sdk/blob/5f9c3fdf88952ea43316f1a18de572e7ae3c13f6/types/context.go) used to keep track of important data during execution of the transaction keeps a `GasMeter` which tracks how much gas has been used.
* `RunTx` returns a result, which `CheckTx` formats into an ABCI `Response` which includes a log, data pertaining to the messages involved, and information about the amount of gas used.

### Gas Checking
If the transaction sender inputted maximum gas previously, it is known as the value `GasWanted`. `CheckTx` returns `GasUsed` which may or may not be less than the user's provided `GasWanted`. A `PostCheckFunc` is an optional filter run after `CheckTx` that can be used to enforce that users provide enough gas.

If at any point during the checking stage the transaction is found to be invalid, the default protocol is to discard it and the transaction's lifecycle ends here. Otherwise, the transaction is a candidate to be included in the next block.

## Consensus
At each round, a proposer is chosen amongst the validators to create and propose the next block. This validator (presumably honest) has generated a Mempool of validated transactions and now includes them in a block. The validators execute [Tendermint BFT Consensus](https://tendermint.com/docs/spec/consensus/consensus.html#terms); with 2/3 approval from the validators, the block is committed. Note that not all blocks have the same number of transactions and it is possible for consensus to result in a nil block or one with no transactions - here, it is assumed that the transaction has made it this far.

## State Changes
During consensus, the validators came to agreement on not only which transactions but also the precise order of the transactions. However, apart from committing to this block in consensus, the ultimate goal is actually for nodes to commit to a new state generated by the transaction state changes.
In order to execute the transaction, the following ABCI function calls are made in order to communicate to the application what state changes need to be made. While nodes each run everything individually, since the messages' state transitions are deterministic and the order was finalized during consensus, this process yields a single, unambiguous result.
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
`BeginBlock` is run first, and mainly transmits important data such as block header and Byzantine Validators from the last round of consensus to be used during the next few steps. No transactions are handled here.

#### DeliverTx
The `DeliverTx` ABCI function defined in [`baseapp`](https://github.com/cosmos/cosmos-sdk/blob/8b1d75caa2099800ee9983e4a4528bcd00fec302/baseapp/baseapp.go
) does the bulk of the state change work: it is run for each transaction in the block in sequential order as committed to during consensus. Under the hood, `DeliverTx` is almost identical to `CheckTx` but calls the [`runTx`](https://github.com/cosmos/cosmos-sdk/blob/cec3065a365f03b86bc629ccb6b275ff5846fdeb/baseapp/baseapp.go#L757-L873) function in deliver mode instead of check mode.

The application utilizes both `AnteHandler` to check and `MsgHandler` to deliver, persisting changes in both `checkTxState` and `deliverTxState`, respectively. This second check happens because nodes may not have seen the same transactions in the same order during the Addition to Mempool stage and a malicious proposer may have included invalid transactions.

`BlockGasMeter` is used to keep track of how much gas is left for each transaction; GasUsed is deducted from it and returned in the Response. Any failed state changes resulting from invalid transactions or `BlockGasMeter` running out causes the transaction processing to terminate and any state changes to revert. Any leftover gas is returned to the user.

#### EndBlock
[`EndBlock`](https://github.com/cosmos/cosmos-sdk/blob/9036430f15c057db0430db6ec7c9072df9e92eb2/baseapp/baseapp.go#L875-L886) is always run at the end and is useful for automatic function calls or changing governance/validator parameters. No transactions are handled here.

#### Commit
The application's `Commit` method is run in order to finalize the state changes made by executing this block's transactions. A new state root should be sent back to serve as a merkle proof for the state change. Any application can inherit Baseapp's [`Commit`](https://github.com/cosmos/cosmos-sdk/blob/cec3065a365f03b86bc629ccb6b275ff5846fdeb/baseapp/baseapp.go#L888-L912) method; it synchronizes all the states by writing the `deliverTxState` into the application's internal state, updating both `checkTxState` and `deliverTxState` afterward.

# Next
Learn about [Baseapp](https://github.com/cosmos/cosmos-sdk/blob/master/docs/concepts/baseapp.md).
