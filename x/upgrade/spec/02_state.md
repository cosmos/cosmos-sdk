<!--
order: 2
-->

# State

The internal state of the `x/upgrade` module is relatively minimal and simple. The
state contains the currently active upgrade `Plan` (if one exists) by key
`0x0` and if a `Plan` is marked as "done" by key `0x1`. The state maintains a 
`Protocol Version` which can be accessed by key `0x3`. 

- Plan: `0x0 -> Plan`
- Done: `0x1 | byte(plan name)  -> BigEndian(Block Height)`
- ProtocolVersion: `0x3 -> BigEndian(Protocol Version)`

The `x/upgrade` module contains no genesis state.
