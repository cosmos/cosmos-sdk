## Deployment

Up until this point, we have only been testing the code as a blockchain with a single validator node running locally.
This is nice for developing, but it's not a real distributed application yet.

This section will demonstrate how to launch your basecoin-based application along 
with a tendermint testnet and initialize the genesis block for fun and profit.
We do this using the [mintnet-kubernetes tool](https://github.com/tendermint/mintnet-kubernetes).
