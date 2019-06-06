# Transaction Lifecycle

## Prerequisite Reading
* ABCI(link)
* Baseapp concept doc (link)


```go				
	        	-----------------------	
	        	|		                  | 
		        |      Creation       |    
User	      |		                  |  
	        	-----------------------		
		                   |		
			                 v			
	        	-----------------------	
        		|		                  | 
		        | Addition to Mempool |
Full-Nodes	|		                  |  
		        -----------------------		
		                   |			
			                 v			
		        -----------------------	
        		|		                  |
	        	|     Consensus	      |  
Validators	|		                  |  
	         	-----------------------	
		                   |			
			                 v			
	      	  -----------------------	
	      	  |		                  |
	      	  |    State Changes    |  
Full-Nodes	|		                  |  
	        	-----------------------	
```

## Definition and Creation
Transactions cause state changes and are comprised of one or more `Msg`s: the developer of an application defines `Msg`s specific to the application. 

## Addition to Mempool
 [Mempool](https://github.com/tendermint/tendermint/blob/75ffa2bf1c7d5805460d941a75112e6a0a38c039/mempool/mempool.go), their pool of transactions approved for inclusion in a block. 
 
 [peer-to-peer communication](https://github.com/tendermint/tendermint/tree/4a568fcedb09493567b293a52c6c42f8d40076c7/p2p).

 [runTx](https://github.com/cosmos/cosmos-sdk/blob/9036430f15c057db0430db6ec7c9072df9e92eb2/baseapp/baseapp.go#L814-L855) 


## Consensus
 [Tendermint BFT Consensus](https://tendermint.com/docs/spec/consensus/consensus.html#terms), with 2/3 approval from the validators, the block is committed.

## State Changes
```
		-----------------------		
		| BeginBlock	        |        
		-----------------------		
		          |		
			        v			
		-----------------------		    
		| DeliverTx(tx0)      |  
		| DeliverTx(tx1)      |   	  
		| DeliverTx(tx2)      |  
		| DeliverTx(tx3)      |  
		|	         .	        |  
		|	         .	        | 
		|          .	        | 
		-----------------------		
		          |			
			        v			
		-----------------------	
		|     EndBlock	      |         
		-----------------------	
		          |			
			        v			
		-----------------------	
		|       Commit	      |         
		-----------------------	
```
#### BeginBlock

#### DeliverTx
[runTx](https://github.com/cosmos/cosmos-sdk/blob/cec3065a365f03b86bc629ccb6b275ff5846fdeb/baseapp/baseapp.go#L757-L873), [runMsgs](https://github.com/cosmos/cosmos-sdk/blob/9036430f15c057db0430db6ec7c9072df9e92eb2/baseapp/baseapp.go#L662-L720), 


#### EndBlock
[`EndBlock`](https://github.com/cosmos/cosmos-sdk/blob/9036430f15c057db0430db6ec7c9072df9e92eb2/baseapp/baseapp.go#L875-L886) 

#### Commit
`Commit` as implemented in [BaseApp](https://github.com/cosmos/cosmos-sdk/blob/cec3065a365f03b86bc629ccb6b275ff5846fdeb/baseapp/baseapp.go#L888-L912).

