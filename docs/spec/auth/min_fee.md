# Minimum Fees specifications

Fees serve two purposes for the operator the network.

Fees rate limit the growth of the state in stored by every full node.and the allow for a general purpose for censorship of transactions that consume resources while being of little economic value. Fees are best suited as an antispam mechanism where validators are disinterested in the use of the network and idenities of users

Node operators should set minimum fees and unit costs of gas in each token denomination they wish to support like.

`--minimum=1steak,5photinos --gas_unit_cost=2steak,4photinos`

When adding transactions to mempool or gossipping transactions, the transactions fee should be check to see if the provided fee equals or exceeds any of the min fee demoninations provided in the configurations.

Tendermint does not currently provide fee based mempool prioritization and fee based mempool filtering is local to node and not part of consensus.

Because the market value for tokens will fluctuate, validators are expected to dynamically adjust the min_fees to a level the encourage use of the network.
