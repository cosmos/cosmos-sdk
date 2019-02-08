# Gas & Fees

Fees serve two purposes for an operator of the network.

Fees limit the growth of the state stored by every full node and allow for
general purpose censorship of transactions of little economic value. Fees
are best suited as an anti-spam mechanism where validators are disinterested in
the use of the network and identities of users.

Fees are determined by the gas limits and gas prices transactions provide, where
`fees = ceil(gasLimit * gasPrices)`. Txs incur gas costs for all state reads/writes,
signature verification, as well as costs proportional to the tx size. Operators
should set minimum gas prices when starting their nodes. They must set the unit
costs of gas in each token denomination they wish to support:

`gaiad start ... --minimum-gas-prices=0.00001steak;0.05photinos`

When adding transactions to mempool or gossipping transactions, validators check
if the transaction's gas prices, which are determined by the provided fees, meet
any of the validator's minimum gas prices. In other words, a transaction must
provide a fee of at least one denomination that matches a validator's minimum
gas price.

Tendermint does not currently provide fee based mempool prioritization, and fee
based mempool filtering is local to node and not part of consensus. But with
minimum gas prices set, such a mechanism could be implemented by node operators.

Because the market value for tokens will fluctuate, validators are expected to
dynamically adjust their minimum gas prices to a level that would encourage the
use of the network.
