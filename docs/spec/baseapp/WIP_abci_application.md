# ABCI application

The `BaseApp` struct fulfills the tendermint-abci `Application` interface. 

## Info

## SetOption

## Query

## CheckTx

## InitChain
TODO pseudo code 

During chain initialization InitChain runs the initialization logic directly on
the CommitMultiStore and commits it. The deliver and check states are
initialized with the ChainID. Additionally the Block gas meter is initialized
with an infinite amount of gas to run any genesis transactions.

Note that we do not `Commit` during `InitChain` however BeginBlock for block 1
starts from this deliverState.


## BeginBlock
TODO complete description & pseudo code 

The block gas meter is reset within BeginBlock for the deliver state. 
If no maximum block gas is set within baseapp then an infinite 
gas meter is set, otherwise a gas meter with the baseapp `maximumBlockGas` 
is initialized 

## DeliverTx
TODO complete description & pseudo code 

Before transaction logic is run, the `BlockGasMeter` is first checked for 
remaining gas. If no gas remains, then `DeliverTx` immediately returns an error. 

After the transaction has been processed the used gas is deducted from the
BlockGasMeter. If the remaining gas exceeds the meter's limits, then DeliverTx
returns an error and the transaction is not committed. 

## EndBlock

## Commit

