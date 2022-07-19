<!--
order: 2
-->

# State

The internal state of the `x/upgrade` module is relatively minimal and simple. The
state contains the currently active upgrade `Plan` (if one exists) by key
`0x0` and if a `Plan` is marked as "done" by key `0x1`. The state
contains the consensus versions of all app modules in the application. The versions
are stored as big endian `uint64`, and can be accessed with prefix `0x2` appended
by the corresponding module name of type `string`. The state maintains a
`Protocol Version` which can be accessed by key `0x3`.

* Plan: `0x0 -> Plan`
* Done: `0x1 | byte(plan name)  -> BigEndian(Block Height)`
* ConsensusVersion: `0x2 | byte(module name)  -> BigEndian(Module Consensus Version)`
* ProtocolVersion: `0x3 -> BigEndian(Protocol Version)`

The `x/upgrade` module contains no genesis state.
