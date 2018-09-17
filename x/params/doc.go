package params

/*
Package params provides a globally available parameter store .

There are two main types, Keeper and Space. Space is an isolated namespace, prefixed by preconfigured spacename. Keeper has a permission to access all existing spaces and create new space.

Space can be used by the individual keepers, who needs a private parameter store that the other keeper are not able to modify. Keeper can be used by the Governance keeper, who need to modify any parameter in case of the proposal passes.

Basic Usage:



*/
