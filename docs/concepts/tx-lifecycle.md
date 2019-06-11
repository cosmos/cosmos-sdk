# Transaction Lifecycle

## Prerequisite Reading
* [High-level overview of the architecture of an SDK application](https://github.com/cosmos/cosmos-sdk/docs/intro/sdk-app-architecture.md)(replace with Anatomy of SDK?)
* Baseapp concept doc?

## Synopsis
1. **Definition and Creation:** Transactions are comprised of metadata and `Msg`s specified by the developer. A user interacts with the Application CLI to call `txCmd`s. 
2. **Addition to Mempool:** All full nodes that receive transactions validate them first by running stateless and stateful checks on a copy of the internal state. Approved transactions are kept in the node's Mempool pending inclusion in the next block.
3. **Consensus:** The proposer creates a block, determining the transactions and their order for this round. Validators run Tendermint BFT Consensus and (if successful) commit to one block.
4. **State Changes:** Transactions are delivered, creating state changes and expending gas. To commit state changes, internal state is updated and all copies are reset; the new state root is returned as proof.

## Definition and Creation
### Definition
**[Transactions](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L36-L43)** are comprised of one or multiple **Messages** and trigger state changes. The developer defines the specific messages in the application module(s) to describe possible actions for the application by implementing the [`Msg`](https://github.com/cosmos/cosmos-sdk/blob/97d10210beb55ad4bd6722f7186a80bf7cb140e2/types/tx_msg.go#L10-L31) interface. He also defines [`Handler`](https://github.com/cosmos/cosmos-sdk/blob/1cfc868d86a152b523443154c8723de832dbec81/types/handler.go#L4)s that execute the actions for each message and return the [`Result`](https://github.com/cosmos/cosmos-sdk/blob/1cfc868d86a152b523443154c8723de832dbec81/types/result.go#L14-L37). An [`AnteHandler`](https://github.com/cosmos/cosmos-sdk/blob/1cfc868d86a152b523443154c8723de832dbec81/types/handler.go#L8) can be defined to execute a message's actions in simulation mode (i.e. without persisting state changes) to perform checks. In the application's Application Command-Line Interface `main.go` file, the developer defines `txCmd` functions that return [Cobra commands](https://godoc.org/github.com/spf13/cobra#Command) to call the application's commands and other lower level functions. 

### Creation
A user can create a transaction by running `appcli tx [tx]` from the command-line, providing transaction data and a value `GasWanted` indicating the maximum amount of gas he is willing to spend to make this action go through. This command directly calls `txCmd`. The node from which this transaction originates broadcasts it to its peers.

## Addition to Mempool
Each full node that receives a transaction performs local checks to filter out invalid transactions before they get included in a block. The transactions approved by a node are held in its [**Mempool**](https://github.com/tendermint/tendermint/blob/75ffa2bf1c7d5805460d941a75112e6a0a38c039/mempool/mempool.go) (memory pool unique to each node) pending approval from the rest of the network. Honest nodes will discard any transactions they find invalid. Prior to consensus, nodes continuously validate incoming transactions and gossip them to their peers.

### Stateless Checks
Upon first receiving the transaction, an optional `PreCheckFunc` can be called to reject transactions that are obviously invalid such as ones exceeding the block size. 
It unwraps the transaction into its message(s) and run each `validateBasic` function, which is simply a stateless sanity check (e.g. nonnegative numbers, nil strings, empty addresses). 

### Stateful Checks
A stateful ABCI validation function, `CheckTx`, is also run: the message's `AnteHandler` performs the actions required for each message using a deep copy of the internal state, `checkTxState`, to validate the transaction without modifying the last committed state. At a given moment, full nodes typically have multiple versions of the application's internal state that may differ from one another; they are synced upon each commit to reflect the new state of the network. The stateful check is able to detect errors such as insufficient funds held by an address or attempted double-spends. 

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
### PostChecks
`CheckTx` also returns `GasUsed` which may or may not be less than the user's provided `GasWanted`. A `PostCheckFunc` is an optional filter after `CheckTx` that can be used to enforce that users must provide enough gas. 


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


