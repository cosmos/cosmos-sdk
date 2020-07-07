<!--
order: 1
-->

# Concepts

## Capabilities

Capabilities are multi-owner. A scoped keeper can create a capability via `NewCapability`
which creates a new unique, unforgeable object-capability reference. The newly
created capability is automatically persisted; the calling module need not call
`ClaimCapability`. Calling `NewCapability` will create the capability with the
calling module and name as a tuple to be treated the capabilities first owner.

Capabilities can be claimed by other modules which add them as owners. `ClaimCapability`
allows a module to claim a capability key which it has received from another
module so that future `GetCapability` calls will succeed. `ClaimCapability` MUST
be called if a module which receives a capability wishes to access it by name in
the future. Again, capabilities are multi-owner, so if multiple modules have a
single Capability reference, they will all own it. If a module receives a capability
from another module but does not call `ClaimCapability`, it may use it in the executing
transaction but will not be able to access it afterwards.

`AuthenticateCapability` can be called by any module to check that a capability
does in fact correspond to a particular name (the name can be un-trusted user input)
with which the calling module previously associated it.

`GetCapability` allows a module to fetch a capability which it has previously
claimed by name. The module is not allowed to retrieve capabilities which it does
not own.

## Stores

- MemStore
