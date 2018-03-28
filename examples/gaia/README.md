Gaiad is the abci application, which can be run stand-alone, or in-process with tendermint.

Gaiacli is a client application, which connects to tendermint rpc, and sends transactions and queries the state. It uses light-client proofs to guarantee the results even if it doesn't have 100% trust in the node it connects to.
