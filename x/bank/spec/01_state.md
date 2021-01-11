<!--
order: 1
-->

# State

The `x/bank` module keeps state of two primary objects, account balances and the
total supply of all balances.

- Balances: `[]byte("balances") | []byte(address) / []byte(balance.Denom) -> ProtocolBuffer(balance)`
- Supply: `0x0 -> ProtocolBuffer(Supply)`
