<!--
order: 1
-->

# State

The `x/bank` module keeps state of the following primary objects:

1. Account balances
2. Denomination metadata
3. The total supply of all balances
4. Information on which denominations are allowed to be sent.

In addition, the `x/bank` module keeps the following indexes to manage the
aforementioned state:

* Supply Index: `0x0 | byte(denom) -> byte(amount)`
* Denom Metadata Index: `0x1 | byte(denom) -> ProtocolBuffer(Metadata)`
* Balances Index: `0x2 | byte(address length) | []byte(address) | []byte(balance.Denom) -> ProtocolBuffer(balance)`
* Reverse Denomination to Address Index: `0x03 | byte(denom) | 0x00 | []byte(address) -> 0`

## Params

The bank module stores it's params in state with the prefix of `0x05`,
it can be updated with governance or the address with authority.

* Params: `0x05 | ProtocolBuffer(Params)`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc3/proto/cosmos/bank/v1beta1/bank.proto#L11-L16
