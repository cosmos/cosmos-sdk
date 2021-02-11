<!--
order: 1
-->

# State

The `x/bank` module keeps state of two primary objects, account balances and the
total supply of all balances.

- Supply: `0x0 -> ProtocolBuffer(Supply)`
- Balances: `0x2 | byte(address length) | []byte(address) | []byte(balance.Denom) -> ProtocolBuffer(balance)`
