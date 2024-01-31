<!--
order: 1
-->

# State

The `x/bank` module keeps state of three primary objects:

1. Account balances
2. Denomination metadata
3. The total supply of all balances

In addition, the `x/bank` module keeps the following indexes to manage the
aforementioned state:

* Supply Index: `0x0 | byte(denom) -> byte(amount)`
* Denom Metadata Index: `0x1 | byte(denom) -> ProtocolBuffer(Metadata)`
* Balances Index: `0x2 | byte(address length) | []byte(address) | []byte(balance.Denom) -> ProtocolBuffer(balance)`
* Reverse Denomination to Address Index: `0x03 | byte(denom) | 0x00 | []byte(address) -> 0`
