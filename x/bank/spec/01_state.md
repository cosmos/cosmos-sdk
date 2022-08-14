<!--
order: 1
-->

# State

The `x/bank` module keeps state of three primary objects, account balances, denom metadata and the
total supply of all balances.

- Supply: `0x0 | byte(denom) -> byte(amount)`
- Denom Metadata: `0x1 | byte(denom) -> ProtocolBuffer(Metadata)`
- Balances: `0x2 | byte(address length) | []byte(address) | []byte(balance.Denom) -> ProtocolBuffer(balance)`
