<!--
order: 2
-->

# State

The internal state of the `x/upgrade` module is relatively minimal and simple. The
state contains the currently active upgrade `Plan` (if one exists) by key
`0x0` and if a `Plan` is marked as "done" by key `0x1`. Additionally, the state
contains the consensus versions of all app modules in the application and their respective
consensus versions. The versions are stored as little endian `uint64`, and can be 
accessed with prefix `0x2` appended by the corresponding module name of type `string`. 

The `x/upgrade` module contains no genesis state.
